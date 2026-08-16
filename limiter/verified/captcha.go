package verified

import (
	"context"
	"errors"
)

type CaptchaVerifier[S comparable] interface {
	Generate(ctx context.Context, driverName string, opts ...Option) (id, question string, err error)
	Verify(ctx context.Context, id, answer string) (bool, error)
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
type Captcha[P CaptchaDriver, B StorageBackend] struct {
	p       P      // captcha provider
	backend B      // store backend
	param   *Param // 验证码参数
}

// NewCaptcha new captcha instance.
func NewCaptcha[P CaptchaDriver, B StorageBackend](p P, backend B, param *Param) *Captcha[P, B] {
	return &Captcha[P, B]{
		p:       p,
		backend: backend,
		param:   param,
	}
}

// Name the provider name
func (c *Captcha[P, B]) Name(driverName string) string {
	return c.p.Driver(driverName).Name()
}

// Generate generate id, question.
func (c *Captcha[P, B]) Generate(ctx context.Context, driverName string, opts ...Option) (id, question string, err error) {
	qa, err := c.p.Driver(driverName).GenerateChallenge(ctx)
	if err != nil {
		return "", "", err
	}
	p := c.param.clone().apply(opts...)
	err = c.backend.Save(ctx, &SaveArgs{
		Key:         p.formatKey(qa.Id),
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
func (c *Captcha[P, B]) Verify(ctx context.Context, id, answer string) (bool, error) {
	return c.backend.Verify(ctx, &VerifyArgs{
		Key:    c.param.formatKey(id),
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
