package verified

import (
	"context"
)

type CaptchaVerifierRegistry[S comparable] interface {
	Generate(ctx context.Context, driverName string, scene S, opts ...Option) (id, question string, err error)
	Verify(ctx context.Context, scene S, id, answer string) (bool, error)
}

// CaptchaRegistry verified captcha limit
type CaptchaRegistry[S comparable, P CaptchaDriver, B StorageBackend] struct {
	p       P            // captcha provider
	backend B            // store backend
	params  map[S]*Param // param set, scene -> 验证码参数
}

// NewCaptchaRegistry new captcha instance.
func NewCaptchaRegistry[S comparable, P CaptchaDriver, B StorageBackend](p P, backend B) *CaptchaRegistry[S, P, B] {
	return &CaptchaRegistry[S, P, B]{
		p:       p,
		backend: backend,
		params:  make(map[S]*Param),
	}
}

func (c *CaptchaRegistry[S, P, B]) SetParams(params map[S]*Param) *CaptchaRegistry[S, P, B] {
	c.params = params
	return c
}

func (c *CaptchaRegistry[S, P, B]) SetParam(scene S, param *Param) *CaptchaRegistry[S, P, B] {
	c.params[scene] = param
	return c
}

func (c *CaptchaRegistry[S, P, B]) getParam(scene S, opts ...Option) (*Param, error) {
	p, ok := c.params[scene]
	if !ok {
		return nil, ErrSceneParamNotFound
	}
	return p.clone().apply(opts...), nil
}

// Name the provider name
func (c *CaptchaRegistry[S, P, B]) Name(driverName string) string {
	return c.p.Driver(driverName).Name()
}

// Generate generate id, question.
func (c *CaptchaRegistry[S, P, B]) Generate(ctx context.Context, driverName string, scene S, opts ...Option) (id, question string, err error) {
	p, err := c.getParam(scene, opts...)
	if err != nil {
		return "", "", err
	}
	return NewCaptcha(c.p, c.backend, p).Generate(ctx, driverName, opts...)
}

// Verify the answer.
func (c *CaptchaRegistry[S, P, B]) Verify(ctx context.Context, scene S, id, answer string) (bool, error) {
	p, err := c.getParam(scene)
	if err != nil {
		return false, err
	}
	return NewCaptcha(c.p, c.backend, p).Verify(ctx, id, answer)
}
