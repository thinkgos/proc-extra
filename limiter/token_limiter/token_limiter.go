package token_limiter

import (
	"context"
	"time"
)

type Rate interface {
	AllowN(ctx context.Context, id string, now time.Time, n int) bool
	Allow(ctx context.Context, id string) bool
}

// TokenLimiter controls how frequently events are allowed to happen with in one second.
type TokenLimiter[B TokenRateBackend] struct {
	backend   TokenRateBackend
	keyPrefix string
	burst     int
	rate      int
}

// NewTokenLimiter returns a new TokenRate that allows events up to rate and permits bursts of at most burst tokens.
func NewTokenLimiter[B TokenRateBackend](backend B, keyPrefix string, rate, burst int) TokenLimiter[B] {
	return TokenLimiter[B]{
		backend:   backend,
		keyPrefix: keyPrefix,
		rate:      rate,
		burst:     burst,
	}
}

func (t TokenLimiter[B]) Allow(ctx context.Context, id string) bool {
	return t.AllowN(ctx, id, time.Now(), 1)
}

func (t TokenLimiter[B]) AllowN(ctx context.Context, id string, now time.Time, n int) bool {
	allow, err := t.backend.AllowN(ctx, &AllowNRequest{
		Key:   t.keyPrefix + id,
		Rate:  t.rate,
		Burst: t.burst,
		Now:   now,
		N:     n,
	})
	return err == nil && allow
}
