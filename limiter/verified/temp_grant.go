package verified

import (
	"context"
)

// TempGrantProvider the temp grant provider
type TempGrantProvider interface {
	Name() string
	GenerateUniqueId() string
}

// TempGrant temp grant verifier
type TempGrant[K comparable, P TempGrantProvider, B StorageBackend] struct {
	p       P            // temp grant provider
	backend B            // store backend
	params  map[K]*Param // param set, kind -> 验证码参数
}

// NewTempGrant new temp grant verifier instance.
func NewTempGrant[K comparable, P TempGrantProvider, B StorageBackend](p P, s B) *TempGrant[K, P, B] {
	return &TempGrant[K, P, B]{
		p:       p,
		backend: s,
		params:  make(map[K]*Param),
	}
}

// Name the provider name
func (r *TempGrant[K, P, B]) Name() string { return r.p.Name() }

func (r *TempGrant[K, P, B]) SetParams(params map[K]*Param) *TempGrant[K, P, B] {
	r.params = params
	return r
}

func (r *TempGrant[K, P, B]) getParam(kind K, opts ...Option) (*Param, error) {
	p, ok := r.params[kind]
	if !ok {
		return nil, ErrParamKindNotFound
	}
	return p.clone().apply(opts...), nil
}

// Issue a temp grant token. use option overwrite default param.
func (r *TempGrant[K, P, B]) Issue(ctx context.Context, kind K, id string, opts ...Option) (string, error) {
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

// Consume the temp grant token.
func (r *TempGrant[K, P, B]) Consume(ctx context.Context, kind K, id, token string) (bool, error) {
	p, err := r.getParam(kind)
	if err != nil {
		return false, err
	}
	return r.backend.Verify(ctx, &VerifyArgs{
		Key:    p.FormatKey(id),
		Answer: token,
	})
}
