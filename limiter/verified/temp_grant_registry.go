package verified

import (
	"context"
)

type TempGranterRegistry[S comparable] interface {
	Issue(ctx context.Context, scene S, id string, opts ...Option) (string, error)
	Consume(ctx context.Context, scene S, id, token string) (bool, error)
}

// TempGrantRegistry temp grant verifier
type TempGrantRegistry[S comparable, P TempGrantGenerator, B StorageBackend] struct {
	p       P            // temp grant provider
	backend B            // store backend
	params  map[S]*Param // param set, scene -> 临时授权码参数
}

// NewTempGrantRegistry new temp grant verifier instance.
func NewTempGrantRegistry[S comparable, P TempGrantGenerator, B StorageBackend](p P, s B) *TempGrantRegistry[S, P, B] {
	return &TempGrantRegistry[S, P, B]{
		p:       p,
		backend: s,
		params:  make(map[S]*Param),
	}
}

// Name the provider name
func (t *TempGrantRegistry[S, P, B]) Name() string { return t.p.Name() }

func (t *TempGrantRegistry[S, P, B]) SetParams(params map[S]*Param) *TempGrantRegistry[S, P, B] {
	t.params = params
	return t
}

func (t *TempGrantRegistry[S, P, B]) SetParam(scene S, param *Param) *TempGrantRegistry[S, P, B] {
	t.params[scene] = param
	return t
}

func (t *TempGrantRegistry[S, P, B]) getParam(scene S, opts ...Option) (*Param, error) {
	p, ok := t.params[scene]
	if !ok {
		return nil, ErrSceneParamNotFound
	}
	return p.clone().apply(opts...), nil
}

// Issue a temp grant token. use option overwrite default param.
func (t *TempGrantRegistry[S, P, B]) Issue(ctx context.Context, scene S, id string, opts ...Option) (string, error) {
	p, err := t.getParam(scene, opts...)
	if err != nil {
		return "", err
	}
	return NewTempGrant(t.p, t.backend, p).Issue(ctx, id, opts...)
}

// Consume the temp grant token.
func (t *TempGrantRegistry[S, P, B]) Consume(ctx context.Context, scene S, id, token string) (bool, error) {
	p, err := t.getParam(scene)
	if err != nil {
		return false, err
	}
	return NewTempGrant(t.p, t.backend, p).Consume(ctx, id, token)
}
