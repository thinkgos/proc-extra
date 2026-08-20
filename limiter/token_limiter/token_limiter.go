package token_limiter

import (
	"context"
	"slices"
	"time"
)

type SceneValuer interface {
	comparable
	Value() string
}

type Rate[S SceneValuer] interface {
	Allow(ctx context.Context, scene S, id string) bool
	AllowN(ctx context.Context, scene S, id string, n int) bool
	TryAllow(ctx context.Context, scene S, id string) (bool, error)
	TryAllowN(ctx context.Context, scene S, id string, n int) (bool, error)
	AllowAt(ctx context.Context, scene S, id string, now time.Time) bool
	AllowNAt(ctx context.Context, scene S, id string, n int, now time.Time) bool
	TryAllowAt(ctx context.Context, scene S, id string, now time.Time) (bool, error)
	TryAllowNAt(ctx context.Context, scene S, id string, n int, now time.Time) (bool, error)
}

type Param struct {
	Rate  int // rate limit in tokens per second
	Burst int // burst size
}

type SceneParam[S SceneValuer] struct {
	scene S
	param *Param
}

// TokenLimiter controls how frequently events are allowed to happen with in one second.
type TokenLimiter[S SceneValuer, B TokenLimiterBackend] struct {
	backend   B               // backend client
	keyPrefix string          // prefix for the key used in the rate limiter
	param     *Param          // general param
	scenes    []SceneParam[S] // scene param.
}

// NewTokenLimiter returns a new TokenRate that allows events up to rate and permits bursts of at most burst tokens.
func NewTokenLimiter[S SceneValuer, B TokenLimiterBackend](backend B) *TokenLimiter[S, B] {
	return &TokenLimiter[S, B]{
		backend:   backend,
		keyPrefix: "token:limit:",
		param: &Param{
			Rate:  100,
			Burst: 200,
		},
		scenes: make([]SceneParam[S], 0),
	}
}

// SetKeyPrefix sets the key prefix.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (v *TokenLimiter[S, B]) SetKeyPrefix(keyPrefix string) *TokenLimiter[S, B] {
	v.keyPrefix = keyPrefix
	return v
}

// SetGeneralParam sets the general param.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (v *TokenLimiter[S, B]) SetGeneralParam(p *Param) *TokenLimiter[S, B] {
	v.param = p
	return v
}

// SetSceneParam sets the param for a specific scene.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (v *TokenLimiter[S, B]) SetSceneParam(scene S, param *Param) *TokenLimiter[S, B] {
	for i := range v.scenes {
		if v.scenes[i].scene == scene {
			v.scenes[i].param = param
			return v
		}
	}
	v.scenes = append(v.scenes, SceneParam[S]{scene: scene, param: param})
	v.scenes = slices.Clone(v.scenes)
	return v
}
func (v *TokenLimiter[S, B]) useScene(scene S) *Param {
	for i := range v.scenes {
		if v.scenes[i].scene == scene {
			return v.scenes[i].param
		}
	}
	return v.param
}

// Allow uses Redis server time.
func (t *TokenLimiter[S, B]) Allow(ctx context.Context, scene S, id string) bool {
	allow, err := t.TryAllowN(ctx, scene, id, 1)
	return err == nil && allow
}

// AllowN uses Redis server time.
func (t *TokenLimiter[S, B]) AllowN(ctx context.Context, scene S, id string, n int) bool {
	allow, err := t.TryAllowN(ctx, scene, id, n)
	return err == nil && allow
}

// TryAllow uses Redis server time.
func (t *TokenLimiter[S, B]) TryAllow(ctx context.Context, scene S, id string) (bool, error) {
	return t.TryAllowN(ctx, scene, id, 1)
}

// TryAllowN uses Redis server time.
func (t *TokenLimiter[S, B]) TryAllowN(ctx context.Context, scene S, id string, n int) (bool, error) {
	return t.TryAllowNAt(ctx, scene, id, n, time.Time{})
}

// AllowAt uses client-supplied time.
func (t *TokenLimiter[S, B]) AllowAt(ctx context.Context, scene S, id string, now time.Time) bool {
	allow, err := t.TryAllowNAt(ctx, scene, id, 1, now)
	return err == nil && allow
}

// AllowNAt uses client-supplied time.
func (t *TokenLimiter[S, B]) AllowNAt(ctx context.Context, scene S, id string, n int, now time.Time) bool {
	allow, err := t.TryAllowNAt(ctx, scene, id, n, now)
	return err == nil && allow
}

// TryAllowAt uses client-supplied time.
func (t *TokenLimiter[S, B]) TryAllowAt(ctx context.Context, scene S, id string, now time.Time) (bool, error) {
	return t.TryAllowNAt(ctx, scene, id, 1, now)
}

// TryAllowNAt uses client-supplied time.
func (t *TokenLimiter[S, B]) TryAllowNAt(ctx context.Context, scene S, id string, n int, now time.Time) (bool, error) {
	p := t.useScene(scene)
	allow, err := t.backend.AllowN(ctx, &AllowNRequest{
		Key:   t.keyPrefix + scene.Value() + ":" + id,
		Rate:  p.Rate,
		Burst: p.Burst,
		Now:   now,
		N:     n,
	})
	if err != nil {
		return false, err
	}
	return allow, nil
}
