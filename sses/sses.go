package sses

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/panjf2000/ants/v2"
)

const heartbeat = ": heartbeat\n\n"

// DefaultEventType is the default event type if not provided in the event
const DefaultEventType = "message"

// Stats stats
type Stats struct {
	ReqSuccess  atomic.Int64 // request success count
	ReqFailure  atomic.Int64 // request failure count
	ReqTimeout  atomic.Int64 // request timeout count
	SendSuccess atomic.Int64 // send success count
	SendFailure atomic.Int64 // send failure count
}

// Store defines the interface for storing and retrieving events
type Store interface {
	// Save the event
	Save(ctx context.Context, channel string, e *Event) error
	// ListByLastId list events by channel, event type and last id, return events and last id, if no more events, return empty slice
	ListByLastId(ctx context.Context, channel, eventType, lastId string, pageSize int) ([]*Event, error)
}

// Option hub option
type Option func(*options)

type options struct {
	store        Store         // 数据存储, 如果设置了, 断开连接后, 会重新发送持久化的事件
	bufferSize   int           // 消息缓冲区大小
	heartbeat    time.Duration // 心跳间隔
	retryLimit   int           // 重试次数
	retryTimeout time.Duration // 重试超时时间
}

func defaultOptions() *options {
	return &options{
		store:        nil,
		bufferSize:   1000,
		heartbeat:    time.Second * 30,
		retryLimit:   3,
		retryTimeout: time.Second * 3,
	}
}

func (o *options) apply(opts ...Option) *options {
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// WithStore set store
func WithStore(store Store) Option {
	return func(h *options) {
		h.store = store
	}
}

// WithBufferSize set message events buffer size
func WithBufferSize(size int) Option {
	return func(h *options) {
		if size > 0 {
			h.bufferSize = size
		}
	}
}

// WithHeartbeat set heartbeat interval
func WithHeartbeat(t time.Duration) Option {
	return func(h *options) {
		if t > 0 {
			h.heartbeat = t
		}
	}
}

// WithRetryLimit set retry limit
func WithRetryLimit(size int) Option {
	return func(h *options) {
		if size > 0 {
			h.retryLimit = size
		}
	}
}

// WithRetryTimeout set retry timeout
func WithRetryTimeout(t time.Duration) Option {
	return func(h *options) {
		if t > 0 {
			h.retryTimeout = t
		}
	}
}

// Hub event center, manage client connections, receive user events, and broadcast them to online users
type Hub struct {
	sessions *SessionManager
	stats    *Stats
	options
}

// NewHub create a new event center
func NewHub(opts ...Option) *Hub {
	o := defaultOptions().apply(opts...)
	return &Hub{
		sessions: NewSessionManager(),
		stats:    &Stats{},
		options:  *o,
	}
}

// SessionTotal get session total
func (h *Hub) SessionTotal() int { return h.sessions.SessionTotal() }

// SessionTotal get session total
func (h *Hub) SessionTotalByChannel(channel string) int {
	return h.sessions.SessionTotalByChannel(channel)
}

// Stats return the stats
func (h *Hub) Stats() *Stats { return h.stats }

// Close event center and stop all worker
func (h *Hub) Close() error { return nil }

// Broadcast events to the channel.
func (h *Hub) Broadcast(ctx context.Context, channel string, events ...*Event) error {
	sessions := h.sessions.Collect(channel)
	for _, e := range events {
		if !isValidEvent(e) {
			continue
		}
		if e.Id == "" {
			e.Id = NewEventId()
		}
		if h.store != nil {
			if err := h.store.Save(ctx, channel, e); err != nil {
				return fmt.Errorf("save event failure, %w", err)
			}
		}
		for _, ses := range sessions {
			h.tryPublish(ctx, ses, e, true)
		}
	}
	return nil
}

// Publish events to specified users who subscribe the channel.
func (h *Hub) Publish(ctx context.Context, channel, userId string, events ...*Event) error {
	return h.publish(ctx, channel, userId, true, events...)
}

// PublishSync events to specified users who subscribe the channel.
func (h *Hub) PublishSync(ctx context.Context, channel, userId string, events ...*Event) error {
	return h.publish(ctx, channel, userId, false, events...)
}

func (h *Hub) publish(ctx context.Context, channel, userId string, async bool, events ...*Event) error {
	sessions := h.sessions.CollectForUser(channel, userId)
	for _, e := range events {
		if !isValidEvent(e) {
			continue
		}
		if e.Id == "" {
			e.Id = NewEventId()
		}
		if h.store != nil {
			if err := h.store.Save(ctx, channel, e); err != nil {
				return fmt.Errorf("save event failure, %w", err)
			}
		}
		for _, ses := range sessions {
			h.tryPublish(ctx, ses, e, async)
		}
	}
	return nil
}
func (h *Hub) PublishSession(ctx context.Context, channel, userId, sessionId string, events ...*Event) error {
	return h.publishSession(ctx, channel, userId, sessionId, true, events...)
}

func (h *Hub) PublishSessionSync(ctx context.Context, channel, userId, sessionId string, events ...*Event) error {
	return h.publishSession(ctx, channel, userId, sessionId, false, events...)
}

func (h *Hub) publishSession(ctx context.Context, channel, userId, sessionId string, async bool, events ...*Event) error {
	for _, e := range events {
		if !isValidEvent(e) {
			continue
		}
		if e.Id == "" {
			e.Id = NewEventId()
		}
		if h.store != nil {
			err := h.store.Save(ctx, channel, e)
			if err != nil {
				return fmt.Errorf("save event failure: %v", err)
			}
		}
		for _, ses := range h.sessions.CollectForUserSession(channel, userId, sessionId) {
			h.tryPublish(ctx, ses, e, async)
		}
	}
	return nil
}

func (h *Hub) tryPublish(ctx context.Context, ses *Session, e *Event, async bool) {
	defer func() {
		if e := recover(); e != nil {
			h.stats.ReqFailure.Add(1)
			slog.ErrorContext(ctx, "tryPublish cause panic", slog.Any("error", e))
		}
	}()
	select {
	case ses.Message <- e:
		h.stats.ReqSuccess.Add(1)
	default:
		if async {
			err := ants.Submit(func() {
				h.tryPushWithTimeout(context.Background(), ses, e)
			})
			if err != nil {
				slog.ErrorContext(ctx, "tryPublish submit task failure", slog.Any("error", err))
			}
		} else {
			h.tryPushWithTimeout(ctx, ses, e)
		}
	}
}

// Asynchronous retry push with timeout logic
func (h *Hub) tryPushWithTimeout(ctx context.Context, ses *Session, e *Event) {
	defer func() {
		if e := recover(); e != nil {
			h.stats.ReqFailure.Add(1)
			slog.ErrorContext(ctx, "tryPushWithTimeout cause panic", slog.Any("error", e))
		}
	}()
	t := time.NewTimer(h.retryTimeout)
	defer t.Stop()
	for range h.retryLimit {
		select {
		case ses.Message <- e:
			h.stats.ReqSuccess.Add(1)
			return
		case <-t.C:
			t.Reset(h.retryTimeout)
		}
	}
	slog.WarnContext(ctx, "push timeout",
		slog.String("userId", ses.UserId),
		slog.String("sessionId", ses.SessionId),
		slog.String("channel", ses.Channel),
		slog.String("eventId", e.Id),
		slog.String("eventType", e.Event),
	)
	h.stats.ReqFailure.Add(1)
	h.stats.ReqTimeout.Add(1)
}

// resend events to specified client after reconnecting
func (h *Hub) resendEvents(ctx context.Context, writer http.ResponseWriter, channel, eventType, lastEventId string) {
	pageSize := 100
	for {
		events, err := h.store.ListByLastId(ctx, channel, eventType, lastEventId, pageSize)
		if err != nil {
			slog.Warn("ListByLastId events error", slog.Any("error", err))
			return
		}
		if len(events) == 0 {
			return
		}
		for _, e := range events {
			lastEventId = e.Id
			if err = e.Render(writer); err != nil {
				slog.Warn("publish event failure!",
					slog.Any("error", err),
					slog.String("eventId", e.Id),
					slog.String("eventType", e.Event),
				)
			}
		}
		if len(events) < pageSize {
			return
		}
	}
}

// isValidEvent checks if the event is valid
func isValidEvent(e *Event) bool {
	return e != nil && e.Event != "" && e.Data != nil
}
