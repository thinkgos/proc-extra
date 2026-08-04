package verified

import (
	"context"
)

// RefluxProvider the reflux provider
type RefluxProvider interface {
	Name() string
	GenerateUniqueId() string
}

// Reflux verified reflux limiter
type Reflux[K comparable, P RefluxProvider, B StorageBackend] struct {
	p       P            // reflux provider
	backend B            // store backend
	params  map[K]*Param // param set, kind -> 验证码参数
}

// NewVerifiedReflux new reflux instance.
func NewVerifiedReflux[K comparable, P RefluxProvider, B StorageBackend](p P, s B) *Reflux[K, P, B] {
	return &Reflux[K, P, B]{
		p:       p,
		backend: s,
		params:  make(map[K]*Param),
	}
}

// Name the provider name
func (r *Reflux[K, P, B]) Name() string { return r.p.Name() }

func (r *Reflux[K, P, B]) SetParams(params map[K]*Param) *Reflux[K, P, B] {
	r.params = params
	return r
}

func (r *Reflux[K, P, B]) getParam(kind K, opts ...Option) (*Param, error) {
	p, ok := r.params[kind]
	if !ok {
		return nil, ErrParamKindNotFound
	}
	return p.clone().apply(opts...), nil
}

// Generate generate uniqueId. use GenerateOption overwrite default key expires
func (r *Reflux[K, P, B]) Generate(ctx context.Context, kind K, id string, opts ...Option) (string, error) {
	p, err := r.getParam(kind, opts...)
	if err != nil {
		return "", err
	}
	answer := r.p.GenerateUniqueId()
	err = r.backend.Save(ctx, &SaveArgs{
		Key:         p.FormatKey(id),
		KeyExpires:  p.KeyExpires,
		MaxAttempts: p.MaxAttempts,
		Answer:      answer,
	})
	if err != nil {
		return "", err
	}
	return answer, nil
}

// Verify the answer.
func (r *Reflux[K, P, B]) Verify(ctx context.Context, kind K, id, answer string) (bool, error) {
	p, err := r.getParam(kind)
	if err != nil {
		return false, err
	}
	return r.backend.Verify(ctx, &VerifyArgs{
		Key:    p.FormatKey(id),
		Answer: answer,
	})
}
