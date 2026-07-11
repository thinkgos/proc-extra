package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/thinkgos/proc-extra/sses"
)

// contextKey is a value for use with context.WithValue. It's used as
// a pointer, so it fits in an interface{} without allocation.
type ctxAuthKey struct{}

// Subject returns the value associated with this context for subjectCtxKey,
func ExtractUserId(w http.ResponseWriter, r *http.Request) string {
	val, _ := r.Context().Value(ctxAuthKey{}).(string)
	return val
}

// WithValueSubject return a copy of parent in which the value associated with
// subjectCtxKey is subject.
func WithValueSubject(ctx context.Context, subject string) context.Context {
	return context.WithValue(ctx, ctxAuthKey{}, subject)
}

func middlewareUserId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := strconv.Itoa(rand.Intn(99) + 100)
		slog.Info("inject userId to context", slog.String("userId", userId))
		r = r.WithContext(WithValueSubject(r.Context(), userId))
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Initialize SSE Hub
	hub := sses.NewHub()
	defer hub.Close() // nolint: errcheck

	// Create Gin router
	r := http.NewServeMux()

	h := hub.Serve(sses.WithServeExtractUserId(ExtractUserId))
	// SSE Event Stream Interface, requires authentication to set uid
	// curl -X GET -H "Content-Type: application/json" http://localhost:8080/sse\?channel\=default
	r.Handle("GET /sse", NewPipeline(middlewareUserId).Handle(h))

	// Register event push endpoint, supports pushing to specified users and broadcast pushing
	// Publish to specified users
	// curl -X POST -H "Content-Type: application/json" -d '{"channel":"default", "userId": "u001", "events":[{"event":"message","data":"hello_publish"}]}' http://localhost:8080/publish
	r.Handle("POST /publish", publishHandler(hub))

	// Broadcast push, not specifying users means pushing to all users
	// curl -X POST -H "Content-Type: application/json" -d '{"channel":"default", "events":[{"event":"message","data":"hello_broadcast"}]}' http://localhost:8080/broadcast
	r.Handle("POST /broadcast", broadcastHandler(hub))

	// Apply a session id
	// curl -X GET -H "Content-Type: application/json" http://localhost:8080/apply
	r.Handle("GET /apply", applySessionIdHandler())

	// simulated event push
	go func() {
		i := 0
		for {
			time.Sleep(time.Second * 5)
			i++
			e := &sses.Event{Event: sses.DefaultEventType, Data: "hello_world_" + strconv.Itoa(i)}
			_ = hub.Broadcast(context.Background(), "default", e) // broadcast push
		}
	}()

	// Start HTTP server
	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}

// PublishRequest push request
type PublishRequest struct {
	Channel string        `json:"channel"`
	UserId  string        `json:"userId"`
	Events  []*sses.Event `json:"events"`
}

func publishHandler(hub *sses.Hub) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := PublishRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responseCode400(w, r, err.Error())
			return
		}
		if err := hub.Publish(r.Context(), req.Channel, req.UserId, req.Events...); err != nil {
			responseCode400(w, r, err.Error())
			return
		}
		responseCode200(w, r, struct{}{})
	})
}

// BroadcastRequest broadcast request
type BroadcastRequest struct {
	Channel string        `json:"channel"`
	Events  []*sses.Event `json:"events"`
}

func broadcastHandler(hub *sses.Hub) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := BroadcastRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responseCode400(w, r, err.Error())
			return
		}
		if err := hub.Broadcast(r.Context(), req.Channel, req.Events...); err != nil {
			responseCode400(w, r, err.Error())
			return
		}
		responseCode200(w, r, struct{}{})
	})
}

// ApplyRequest apply request
type ApplySessionIdReply struct {
	SessionId string `json:"sessionId"`
}

func applySessionIdHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseCode200(w, r, ApplySessionIdReply{SessionId: sses.NewSessionId()})
	})
}

func responseCode400(w http.ResponseWriter, r *http.Request, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]any{"code": http.StatusBadRequest, "msg": msg, "data": struct{}{}})
}

func responseCode200(w http.ResponseWriter, r *http.Request, data any) {
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "msg": "ok", "data": data})
}

type Middleware func(http.Handler) http.Handler

type Pipeline struct {
	middlewares []Middleware
}

func NewPipeline(ms ...Middleware) Pipeline {
	return Pipeline{middlewares: ms}
}
func (p Pipeline) Handle(final http.Handler) http.Handler {
	if len(p.middlewares) == 0 {
		return final
	}
	handle := final
	for i := len(p.middlewares) - 1; i >= 0; i-- {
		handle = p.middlewares[i](handle)
	}
	return handle
}

func (p Pipeline) HandleFunc(final http.HandlerFunc) http.Handler {
	return p.Handle(http.HandlerFunc(final))
}
