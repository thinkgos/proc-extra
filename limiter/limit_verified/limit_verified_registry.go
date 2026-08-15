package limit_verified

import (
	"context"
)

type LimitVerifierRegistry[S comparable] interface {
	Name() string
	SendCode(ctx context.Context, scene S, target, code string) (*SendCodeResult, error)
	VerifyCode(ctx context.Context, scene S, target, code string) (*VerifyResult, error)
}

// LimitVerifiedRegistry limit verified code
type LimitVerifiedRegistry[S comparable, P LimitVerifiedProvider, B LimitVerifiedBackend] struct {
	p         LimitVerifiedProvider // LimitVerifiedProvider send code
	backend   LimitVerifiedBackend  // backend client
	keyPrefix string                // key prefix
	params    map[S]*Param
}

// NewLimitVerifiedRegistry  new a limit verified
func NewLimitVerifiedRegistry[S comparable, P LimitVerifiedProvider, B LimitVerifiedBackend](p P, backend B) *LimitVerifiedRegistry[S, P, B] {
	v := &LimitVerifiedRegistry[S, P, B]{
		p:         p,
		backend:   backend,
		keyPrefix: "limit:verifier:scene:",
		params:    make(map[S]*Param),
	}
	return v
}
func (v *LimitVerifiedRegistry[S, P, B]) SetKeyPrefix(keyPrefix string) *LimitVerifiedRegistry[S, P, B] {
	v.keyPrefix = keyPrefix
	return v
}

func (v *LimitVerifiedRegistry[S, P, B]) SetParams(params map[S]*Param) *LimitVerifiedRegistry[S, P, B] {
	v.params = params
	return v
}
func (v *LimitVerifiedRegistry[S, P, B]) SetParam(scene S, param *Param) *LimitVerifiedRegistry[S, P, B] {
	v.params[scene] = param
	return v
}
func (v *LimitVerifiedRegistry[S, P, B]) getParam(scene S) (*Param, error) {
	param, ok := v.params[scene]
	if !ok {
		return nil, ErrSceneParamNotFound
	}
	return param, nil
}

// Name the provider name
func (v *LimitVerifiedRegistry[S, P, B]) Name() string { return v.p.Name() }

// SendCode send code and backend.
func (v *LimitVerifiedRegistry[S, P, B]) SendCode(ctx context.Context, scene S, target, code string) (*EvaluateResult, error) {
	p, err := v.getParam(scene)
	if err != nil {
		return nil, err
	}
	return NewLimitVerified(v.p, v.backend, p).
		SetKeyPrefix(v.keyPrefix).
		SendCode(ctx, target, code)
}

// VerifyCode verify code from cache.
func (v *LimitVerifiedRegistry[S, P, B]) VerifyCode(ctx context.Context, scene S, target, code string) (*VerifyResult, error) {
	p, err := v.getParam(scene)
	if err != nil {
		return nil, err
	}
	return NewLimitVerified(v.p, v.backend, p).
		SetKeyPrefix(v.keyPrefix).
		VerifyCode(ctx, target, code)
}
