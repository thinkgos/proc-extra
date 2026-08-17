package token_limiter

import (
	"context"
	"errors"
	"maps"
	"time"
)

// ErrSceneParamNotFound is an error that scene's param not found.
var ErrSceneParamNotFound = errors.New("token_limiter: the scene's param not found")

type RateRegistry[S comparable] interface {
	Allow(ctx context.Context, scene S, id string) bool
	AllowN(ctx context.Context, scene S, id string, n int) bool
	TryAllow(ctx context.Context, scene S, id string) (bool, error)
	TryAllowN(ctx context.Context, scene S, id string, n int) (bool, error)
	AllowAt(ctx context.Context, scene S, id string, now time.Time) bool
	AllowNAt(ctx context.Context, scene S, id string, n int, now time.Time) bool
	TryAllowAt(ctx context.Context, scene S, id string, now time.Time) (bool, error)
	TryAllowNAt(ctx context.Context, scene S, id string, n int, now time.Time) (bool, error)
}

// TokenLimiterRegistry controls how frequently events are allowed to happen with in one second.
type TokenLimiterRegistry[S comparable, B TokenLimiterBackend] struct {
	backend B
	params  map[S]*Param
}

// NewTokenLimiterRegistry returns a new TokenRate that allows events up to rate and permits bursts of at most burst tokens.
func NewTokenLimiterRegistry[S comparable, B TokenLimiterBackend](backend B) *TokenLimiterRegistry[S, B] {
	return &TokenLimiterRegistry[S, B]{
		backend: backend,
		params:  map[S]*Param{},
	}
}

// RegistersParams sets all scene params at once.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (c *TokenLimiterRegistry[S, B]) RegistersParams(params map[S]*Param) *TokenLimiterRegistry[S, B] {
	maps.Copy(c.params, params)
	return c
}

// RegisterParam sets the param for a specific scene.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (c *TokenLimiterRegistry[S, B]) RegisterParam(scene S, param *Param) *TokenLimiterRegistry[S, B] {
	c.params[scene] = param
	return c
}

func (c *TokenLimiterRegistry[S, B]) getParam(scene S) (*Param, error) {
	p, ok := c.params[scene]
	if !ok {
		return nil, ErrSceneParamNotFound
	}
	return p, nil
}

// Allow uses Redis server time.
func (t *TokenLimiterRegistry[S, B]) Allow(ctx context.Context, scene S, id string) bool {
	p, err := t.getParam(scene)
	if err != nil {
		return false
	}
	return NewTokenLimiter(t.backend, p).Allow(ctx, id)
}

// AllowN uses Redis server time.
func (t *TokenLimiterRegistry[S, B]) AllowN(ctx context.Context, scene S, id string, n int) bool {
	p, err := t.getParam(scene)
	if err != nil {
		return false
	}
	return NewTokenLimiter(t.backend, p).AllowN(ctx, id, n)
}

// TryAllow uses Redis server time.
func (t *TokenLimiterRegistry[S, B]) TryAllow(ctx context.Context, scene S, id string) (bool, error) {
	p, err := t.getParam(scene)
	if err != nil {
		return false, err
	}
	return NewTokenLimiter(t.backend, p).TryAllow(ctx, id)
}

// TryAllowN uses Redis server time.
func (t *TokenLimiterRegistry[S, B]) TryAllowN(ctx context.Context, scene S, id string, n int) (bool, error) {
	p, err := t.getParam(scene)
	if err != nil {
		return false, err
	}
	return NewTokenLimiter(t.backend, p).TryAllowN(ctx, id, n)
}

// AllowAt uses client-supplied time.
func (t *TokenLimiterRegistry[S, B]) AllowAt(ctx context.Context, scene S, id string, now time.Time) bool {
	p, err := t.getParam(scene)
	if err != nil {
		return false
	}
	return NewTokenLimiter(t.backend, p).AllowAt(ctx, id, now)
}

// AllowNAt uses client-supplied time.
func (t *TokenLimiterRegistry[S, B]) AllowNAt(ctx context.Context, scene S, id string, n int, now time.Time) bool {
	p, err := t.getParam(scene)
	if err != nil {
		return false
	}
	return NewTokenLimiter(t.backend, p).AllowNAt(ctx, id, n, now)
}

// TryAllowAt uses client-supplied time.
func (t *TokenLimiterRegistry[S, B]) TryAllowAt(ctx context.Context, scene S, id string, now time.Time) (bool, error) {
	p, err := t.getParam(scene)
	if err != nil {
		return false, err
	}
	return NewTokenLimiter(t.backend, p).TryAllowAt(ctx, id, now)
}

// TryAllowNAt uses client-supplied time.
func (t *TokenLimiterRegistry[S, B]) TryAllowNAt(ctx context.Context, scene S, id string, n int, now time.Time) (bool, error) {
	p, err := t.getParam(scene)
	if err != nil {
		return false, err
	}
	return NewTokenLimiter(t.backend, p).TryAllowNAt(ctx, id, n, now)
}
