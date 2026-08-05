package verified

import (
	"context"
)

type TempGranter[S comparable] interface {
	Issue(ctx context.Context, scene S, id string, opts ...Option) (string, error)
	Consume(ctx context.Context, scene S, id, token string) (bool, error)
}

// TempGrantGenerator the temp grant generator
type TempGrantGenerator interface {
	Name() string
	GenerateUniqueId() string
}

// TempGrant temp grant verifier
type TempGrant[S comparable, P TempGrantGenerator, B StorageBackend] struct {
	p       P            // temp grant provider
	backend B            // store backend
	params  map[S]*Param // param set, scene -> 临时授权码参数
}

// NewTempGrant new temp grant verifier instance.
func NewTempGrant[S comparable, P TempGrantGenerator, B StorageBackend](p P, s B) *TempGrant[S, P, B] {
	return &TempGrant[S, P, B]{
		p:       p,
		backend: s,
		params:  make(map[S]*Param),
	}
}

// Name the provider name
func (r *TempGrant[S, P, B]) Name() string { return r.p.Name() }

func (r *TempGrant[S, P, B]) SetParams(params map[S]*Param) *TempGrant[S, P, B] {
	r.params = params
	return r
}

func (r *TempGrant[S, P, B]) getParam(scene S, opts ...Option) (*Param, error) {
	p, ok := r.params[scene]
	if !ok {
		return nil, ErrParamKindNotFound
	}
	return p.clone().apply(opts...), nil
}

// Issue a temp grant token. use option overwrite default param.
func (r *TempGrant[S, P, B]) Issue(ctx context.Context, scene S, id string, opts ...Option) (string, error) {
	p, err := r.getParam(scene, opts...)
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
func (r *TempGrant[S, P, B]) Consume(ctx context.Context, scene S, id, token string) (bool, error) {
	p, err := r.getParam(scene)
	if err != nil {
		return false, err
	}
	return r.backend.Verify(ctx, &VerifyArgs{
		Key:    p.FormatKey(id),
		Answer: token,
	})
}
