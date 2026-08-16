package token_limiter

import (
	"context"
	"time"
)

type Rate interface {
	AllowN(ctx context.Context, id string, now time.Time, n int) bool
	Allow(ctx context.Context, id string) bool
	TryAllow(ctx context.Context, id string) (bool, error)
	TryAllowN(ctx context.Context, id string, now time.Time, n int) (bool, error)
}

type Param struct {
	KeyPrefix string // prefix for the key used in the rate limiter
	Rate      int    // rate limit in tokens per second
	Burst     int    // burst size
}

// TokenLimiter controls how frequently events are allowed to happen with in one second.
type TokenLimiter[B TokenLimiterBackend] struct {
	backend B
	param   *Param
}

// NewTokenLimiter returns a new TokenRate that allows events up to rate and permits bursts of at most burst tokens.
func NewTokenLimiter[B TokenLimiterBackend](backend B, param *Param) *TokenLimiter[B] {
	return &TokenLimiter[B]{
		backend: backend,
		param:   param,
	}
}

func (t *TokenLimiter[B]) Allow(ctx context.Context, id string) bool {
	return t.AllowN(ctx, id, time.Now(), 1)
}

func (t *TokenLimiter[B]) AllowN(ctx context.Context, id string, now time.Time, n int) bool {
	allow, err := t.TryAllowN(ctx, id, now, n)
	return err == nil && allow
}

func (t *TokenLimiter[B]) TryAllow(ctx context.Context, id string) (bool, error) {
	return t.TryAllowN(ctx, id, time.Now(), 1)
}

func (t *TokenLimiter[B]) TryAllowN(ctx context.Context, id string, now time.Time, n int) (bool, error) {
	allow, err := t.backend.AllowN(ctx, &AllowNRequest{
		Key:   t.param.KeyPrefix + id,
		Rate:  t.param.Rate,
		Burst: t.param.Burst,
		Now:   now,
		N:     n,
	})
	if err != nil {
		return false, err
	}
	return allow, nil
}
