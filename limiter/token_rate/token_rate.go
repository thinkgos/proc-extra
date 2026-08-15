package token_rate

import (
	"context"
	"time"
)

type Rate interface {
	AllowN(now time.Time, n int) bool
	Allow() bool
}

// TokenRate controls how frequently events are allowed to happen with in one second.
type TokenRate[B TokenRateBackend] struct {
	backend TokenRateBackend
	key     string
	burst   int
	rate    int
}

// NewTokenRate returns a new TokenRate that allows events up to rate and permits bursts of at most burst tokens.
func NewTokenRate[B TokenRateBackend](backend B, key string, rate, burst int) *TokenRate[B] {
	return &TokenRate[B]{
		backend: backend,
		key:     key,
		rate:    rate,
		burst:   burst,
	}
}

func (t *TokenRate[B]) Allow(ctx context.Context) bool {
	return t.AllowN(ctx, time.Now(), 1)
}

func (t *TokenRate[B]) AllowN(ctx context.Context, now time.Time, n int) bool {
	allow, err := t.backend.AllowN(ctx, &AllowNRequest{
		Key:   t.key,
		Rate:  t.rate,
		Burst: t.burst,
		Now:   now,
		N:     n,
	})
	return err == nil && allow
}
