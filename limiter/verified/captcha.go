package verified

import (
	"context"
	"errors"
)

type CaptchaVerifier[S comparable] interface {
	Generate(ctx context.Context, driverName string, scene S, opts ...Option) (id, question string, err error)
	Verify(ctx context.Context, scene S, id, answer string) (bool, error)
}

// Challenge question and answer.
type Challenge struct {
	Id       string
	Question string
	Answer   string
}

// ChallengeProvider the captcha challenge provider
type ChallengeProvider interface {
	Name() string
	GenerateChallenge(ctx context.Context) (*Challenge, error)
}

// CaptchaDriver the captcha driver
type CaptchaDriver interface {
	Driver(dName string) ChallengeProvider
}

// Captcha verified captcha limit
type Captcha[S comparable, P CaptchaDriver, B StorageBackend] struct {
	p       P            // captcha provider
	backend B            // store backend
	params  map[S]*Param // param set, scene -> 验证码参数
}

// NewCaptcha new captcha instance.
func NewCaptcha[S comparable, P CaptchaDriver, B StorageBackend](p P, backend B) *Captcha[S, P, B] {
	return &Captcha[S, P, B]{
		p:       p,
		backend: backend,
		params:  make(map[S]*Param),
	}
}

func (c *Captcha[S, P, B]) SetParams(params map[S]*Param) *Captcha[S, P, B] {
	c.params = params
	return c
}

func (c *Captcha[S, P, B]) getParam(scene S, opts ...Option) (*Param, error) {
	p, ok := c.params[scene]
	if !ok {
		return nil, ErrSceneParamNotFound
	}
	return p.clone().apply(opts...), nil
}

// Name the provider name
func (c *Captcha[S, P, B]) Name(driverName string) string {
	return c.p.Driver(driverName).Name()
}

// Generate generate id, question.
func (c *Captcha[S, P, B]) Generate(ctx context.Context, driverName string, scene S, opts ...Option) (id, question string, err error) {
	p, err := c.getParam(scene, opts...)
	if err != nil {
		return "", "", err
	}
	qa, err := c.p.Driver(driverName).GenerateChallenge(ctx)
	if err != nil {
		return "", "", err
	}
	err = c.backend.Save(ctx, &SaveArgs{
		Key:         p.FormatKey(qa.Id),
		KeyExpires:  p.KeyExpires,
		MaxAttempts: p.MaxAttempts,
		Answer:      qa.Answer,
	})
	if err != nil {
		return "", "", err
	}
	return qa.Id, qa.Question, nil
}

// Verify the answer.
func (c *Captcha[S, P, B]) Verify(ctx context.Context, scene S, id, answer string) (bool, error) {
	p, err := c.getParam(scene)
	if err != nil {
		return false, err
	}
	return c.backend.Verify(ctx, &VerifyArgs{
		Key:    p.FormatKey(id),
		Answer: answer,
	})
}

type UnsupportedChallengeProvider struct{}

func (x UnsupportedChallengeProvider) Name() string {
	return "Unsupported captcha driver"
}
func (x UnsupportedChallengeProvider) GenerateChallenge(ctx context.Context) (*Challenge, error) {
	return nil, errors.New(x.Name())
}
