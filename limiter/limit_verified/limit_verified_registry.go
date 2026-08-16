package limit_verified

import (
	"context"
	"maps"
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

// SetKeyPrefix sets the key prefix for the limit verified registry.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (r *LimitVerifiedRegistry[S, P, B]) SetKeyPrefix(keyPrefix string) *LimitVerifiedRegistry[S, P, B] {
	r.keyPrefix = keyPrefix
	return r
}

// RegistersParams sets all scene params at once.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (r *LimitVerifiedRegistry[S, P, B]) RegistersParams(params map[S]*Param) *LimitVerifiedRegistry[S, P, B] {
	maps.Copy(r.params, params)
	return r
}

// RegisterParam sets the param for a specific scene.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (r *LimitVerifiedRegistry[S, P, B]) RegisterParam(scene S, param *Param) *LimitVerifiedRegistry[S, P, B] {
	r.params[scene] = param
	return r
}
func (r *LimitVerifiedRegistry[S, P, B]) getParam(scene S) (*Param, error) {
	param, ok := r.params[scene]
	if !ok {
		return nil, ErrSceneParamNotFound
	}
	return param, nil
}

// Name the provider name
func (r *LimitVerifiedRegistry[S, P, B]) Name() string { return r.p.Name() }

// SendCode send code and backend.
func (r *LimitVerifiedRegistry[S, P, B]) SendCode(ctx context.Context, scene S, target, code string) (*EvaluateResult, error) {
	p, err := r.getParam(scene)
	if err != nil {
		return nil, err
	}
	return NewLimitVerified(r.p, r.backend, p).
		SetKeyPrefix(r.keyPrefix).
		SendCode(ctx, target, code)
}

// VerifyCode verify code from cache.
func (r *LimitVerifiedRegistry[S, P, B]) VerifyCode(ctx context.Context, scene S, target, code string) (*VerifyResult, error) {
	p, err := r.getParam(scene)
	if err != nil {
		return nil, err
	}
	return NewLimitVerified(r.p, r.backend, p).
		SetKeyPrefix(r.keyPrefix).
		VerifyCode(ctx, target, code)
}
