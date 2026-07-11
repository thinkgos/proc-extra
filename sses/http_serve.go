package sses

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/thinkgos/proc/lookup"
)

const CtxUserIdKey = "http/sse/user-id"

// ServeOption is an option for Serve
type ServeOption func(*serveOptions)

type serveOptions struct {
	headers            map[string]string
	extractUserId      func(http.ResponseWriter, *http.Request) string
	extractSessionId   *lookup.Lookup
	extractChannel     *lookup.Lookup
	extractEventType   *lookup.Lookup
	extractLastEventId *lookup.Lookup
	errFallback        func(http.ResponseWriter, *http.Request, error)
	onRegister         func(*Session) // 注册时
	onDeregister       func(*Session) // 注销时
}

func defaultServeOptions() *serveOptions {
	return &serveOptions{
		extractUserId:      func(http.ResponseWriter, *http.Request) string { return "" },
		extractSessionId:   lookup.NewLookup("query:sessionId"),
		extractChannel:     lookup.NewLookup("query:channel"),
		extractEventType:   lookup.NewLookup("query:eventType,header:Event-Type"),
		extractLastEventId: lookup.NewLookup("query:lastEventId,header:Last-Event-ID"),
		errFallback: func(w http.ResponseWriter, r *http.Request, err error) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
		},
		onRegister:   func(s *Session) {},
		onDeregister: func(s *Session) {},
	}
}

func (o *serveOptions) apply(opts ...ServeOption) *serveOptions {
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// WithServeHeaders sets extra headers to be sent with the response.
func WithServeHeaders(headers map[string]string) ServeOption {
	return func(o *serveOptions) {
		o.headers = headers
	}
}

// WithServeExtractUserId(Require) sets the function to extract the user Id from the request.
func WithServeExtractUserId(f func(http.ResponseWriter, *http.Request) string) ServeOption {
	return func(o *serveOptions) {
		o.extractUserId = f
	}
}

// WithServeExtractSessionId sets the function to extract the session Id from the request.
func WithServeExtractSessionId(l *lookup.Lookup) ServeOption {
	return func(o *serveOptions) {
		o.extractSessionId = l
	}
}

// WithServeExtractChannel sets the function to extract the channel from the request.
func WithServeExtractChannel(l *lookup.Lookup) ServeOption {
	return func(o *serveOptions) {
		o.extractChannel = l
	}
}

// WithServeExtractEventType sets the function to extract the event type from the request.
func WithServeExtractEventType(l *lookup.Lookup) ServeOption {
	return func(o *serveOptions) {
		o.extractEventType = l
	}
}

// WithServeExtractLastEventId sets the function to extract the last event id from the request.
func WithServeExtractLastEventId(l *lookup.Lookup) ServeOption {
	return func(o *serveOptions) {
		o.extractLastEventId = l
	}
}

// WithErrorFallback set the fallback handler when request are error happened.
// default: the 400 bad request error to the client
func WithErrorFallback(fn func(http.ResponseWriter, *http.Request, error)) ServeOption {
	return func(o *serveOptions) {
		if fn != nil {
			o.errFallback = fn
		}
	}
}

// WithServeOnRegister sets the function to be called when a user session is registered.
func WithServeOnRegister(fn func(*Session)) ServeOption {
	return func(o *serveOptions) {
		o.onRegister = fn
	}
}

// WithServeOnDeregister sets the function to be called when a user session is deregistered.
func WithServeOnDeregister(fn func(*Session)) ServeOption {
	return func(o *serveOptions) {
		o.onDeregister = fn
	}
}

// Serve serves a client connection
func (h *Hub) Serve(opts ...ServeOption) http.Handler {
	opt := defaultServeOptions().apply(opts...)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		for k, v := range opt.headers {
			w.Header().Set(k, v)
		}

		_, ok := w.(http.Flusher)
		if !ok {
			opt.errFallback(w, r, errors.New("streaming unsupported"))
			return
		}
		//* 获取用户id
		userId := opt.extractUserId(w, r)
		if userId == "" {
			opt.errFallback(w, r, errors.New("userId is empty, not allow connection"))
			return
		}
		//* 获取订阅的channel
		channel, err := opt.extractChannel.ExtractValue(r)
		if err != nil {
			opt.errFallback(w, r, errors.New("channel is empty, not allow connection"))
			return
		}
		//* 获取会话id, 如果没有, 则创建一个
		sessionId := opt.extractSessionId.ExtractValueOr(r, "")
		if sessionId == "" {
			sessionId = NewSessionId()
		}
		//* 获取请求事件类型和最后的事件id
		eventType := opt.extractEventType.ExtractValueOr(r, DefaultEventType)
		lastEventId := opt.extractLastEventId.ExtractValueOr(r, "")
		//* 创建用户会话
		session := &Session{
			UserId:    userId,
			SessionId: sessionId,
			Channel:   channel,
			Message:   make(chan *Event, h.bufferSize),
		}

		//* 注册用户会话
		h.sessions.Add(session)
		// h.logger.OnInfoContext(c.Request.Context()).
		// 	String("channel", session.Channel).
		// 	String("userId", session.UserId).
		// 	String("sessionId", session.SessionId).
		// 	Msg("user session connected")
		opt.onRegister(session)
		defer func() {
			//* 注销用户会话
			opt.onDeregister(session)
			h.sessions.Delete(session)
			// h.logger.OnInfoContext(c.Request.Context()).
			// 	String("userId", session.UserId).
			// 	String("sessionId", session.SessionId).
			// 	String("channel", session.Channel).
			// 	Msg("user session disconnected")
		}()

		//* 重发旧的消息
		if h.store != nil && lastEventId != "" {
			h.resendEvents(ctx, w, channel, eventType, lastEventId)
			w.(http.Flusher).Flush()
		}
		// 会话连接成功, 发送一次心跳, 客户端连接成功
		_, _ = w.Write([]byte(heartbeat))
		w.(http.Flusher).Flush()

		t := time.NewTimer(h.heartbeat)
		defer t.Stop()

		step := func(io.Writer) bool {
			defer t.Reset(h.heartbeat)
			select {
			case e, ok := <-session.Message:
				if !ok { // 关闭
					return false
				}
				err := e.Render(w)
				if err != nil {
					h.stats.SendSuccess.Add(1)
					return false
				} else {
					h.stats.SendFailure.Add(1)
					return true
				}
			case <-t.C:
				// 发送心跳
				_, err := w.Write([]byte(heartbeat))
				return err == nil
			case <-r.Context().Done():
				return false
			}
		}

		for {
			keepOpen := step(w)
			w.(http.Flusher).Flush()
			if !keepOpen {
				return
			}
		}
	})
}
