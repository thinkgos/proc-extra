package verified

import (
	"context"
)

type TempGranter interface {
	Issue(ctx context.Context, id string, opts ...Option) (string, error)
	Consume(ctx context.Context, id, token string) (bool, error)
}

// TempGrantGenerator the temp grant generator
type TempGrantGenerator interface {
	Name() string
	GenerateUniqueId() string
}

// TempGrant temp grant verifier
type TempGrant[P TempGrantGenerator, B StorageBackend] struct {
	p       P      // temp grant provider
	backend B      // store backend
	param   *Param // 临时授权码参数
}

// NewTempGrant new temp grant verifier instance.
//
//go:inline
func NewTempGrant[P TempGrantGenerator, B StorageBackend](p P, b B, param *Param) *TempGrant[P, B] {
	return &TempGrant[P, B]{
		p:       p,
		backend: b,
		param:   param,
	}
}

// Name the provider name
func (t *TempGrant[P, B]) Name() string { return t.p.Name() }

// Issue a temp grant token. use option overwrite default param.
func (t *TempGrant[P, B]) Issue(ctx context.Context, id string, opts ...Option) (string, error) {
	p := t.param.clone().apply(opts...)
	answer := t.p.GenerateUniqueId()
	err := t.backend.Save(ctx, &SaveArgs{
		Key:         p.formatKey(id),
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
func (t *TempGrant[P, B]) Consume(ctx context.Context, id, token string) (bool, error) {
	return t.backend.Verify(ctx, &VerifyArgs{
		Key:    t.param.formatKey(id),
		Answer: token,
	})
}
