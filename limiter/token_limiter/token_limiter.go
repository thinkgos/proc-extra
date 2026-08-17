package token_limiter

import (
	"context"
	"time"
)

type Rate interface {
	Allow(ctx context.Context, id string) bool
	AllowN(ctx context.Context, id string, n int) bool
	TryAllow(ctx context.Context, id string) (bool, error)
	TryAllowN(ctx context.Context, id string, n int) (bool, error)
	AllowAt(ctx context.Context, id string, now time.Time) bool
	AllowNAt(ctx context.Context, id string, n int, now time.Time) bool
	TryAllowAt(ctx context.Context, id string, now time.Time) (bool, error)
	TryAllowNAt(ctx context.Context, id string, n int, now time.Time) (bool, error)
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

// Allow uses Redis server time.
func (t *TokenLimiter[B]) Allow(ctx context.Context, id string) bool {
	return t.AllowN(ctx, id, 1)
}

// AllowN uses Redis server time.
func (t *TokenLimiter[B]) AllowN(ctx context.Context, id string, n int) bool {
	allow, err := t.TryAllowN(ctx, id, n)
	return err == nil && allow
}

// TryAllow uses Redis server time.
func (t *TokenLimiter[B]) TryAllow(ctx context.Context, id string) (bool, error) {
	return t.TryAllowN(ctx, id, 1)
}

// TryAllowN uses Redis server time.
func (t *TokenLimiter[B]) TryAllowN(ctx context.Context, id string, n int) (bool, error) {
	return t.allowN(ctx, id, time.Time{}, n)
}

// AllowAt uses client-supplied time.
func (t *TokenLimiter[B]) AllowAt(ctx context.Context, id string, now time.Time) bool {
	return t.AllowNAt(ctx, id, 1, now)
}

// AllowNAt uses client-supplied time.
func (t *TokenLimiter[B]) AllowNAt(ctx context.Context, id string, n int, now time.Time) bool {
	allow, err := t.TryAllowNAt(ctx, id, n, now)
	return err == nil && allow
}

// TryAllowAt uses client-supplied time.
func (t *TokenLimiter[B]) TryAllowAt(ctx context.Context, id string, now time.Time) (bool, error) {
	return t.TryAllowNAt(ctx, id, 1, now)
}

// TryAllowNAt uses client-supplied time.
func (t *TokenLimiter[B]) TryAllowNAt(ctx context.Context, id string, n int, now time.Time) (bool, error) {
	return t.allowN(ctx, id, now, n)
}

func (t *TokenLimiter[B]) allowN(ctx context.Context, id string, now time.Time, n int) (bool, error) {
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
