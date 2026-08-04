package verified

import (
	"context"
	"errors"
)

// Challenge question and answer for CaptchaDriver driver
type Challenge struct {
	Id       string
	Question string
	Answer   string
}

// CaptchaDriver the captcha driver
type CaptchaDriver interface {
	Name() string
	GenerateQuestionAnswer() (*Challenge, error)
}

// CaptchaProvider the captcha provider
type CaptchaProvider interface {
	AcquireDriver(dName string) CaptchaDriver
}

// Captcha verified captcha limit
type Captcha[K comparable, P CaptchaProvider, B StorageBackend] struct {
	p       P            // captcha provider
	backend B            // store backend
	params  map[K]*Param // param set, kind -> 验证码参数
}

// NewCaptchaVerified new captcha instance.
func NewCaptchaVerified[K comparable, P CaptchaProvider, B StorageBackend](p P, backend B) *Captcha[K, P, B] {
	return &Captcha[K, P, B]{
		p:       p,
		backend: backend,
		params:  make(map[K]*Param),
	}
}

func (c *Captcha[K, P, B]) SetParams(params map[K]*Param) *Captcha[K, P, B] {
	c.params = params
	return c
}

func (c *Captcha[K, P, B]) getParam(kind K, opts ...Option) (*Param, error) {
	p, ok := c.params[kind]
	if !ok {
		return nil, ErrParamKindNotFound
	}
	return p.clone().apply(opts...), nil
}

// Name the provider name
func (c *Captcha[K, P, B]) Name(driverName string) string {
	return c.p.AcquireDriver(driverName).Name()
}

// Generate generate id, question.
func (c *Captcha[K, P, B]) Generate(ctx context.Context, driverName string, kind K, opts ...Option) (id, question string, err error) {
	p, err := c.getParam(kind, opts...)
	if err != nil {
		return "", "", err
	}
	qa, err := c.p.AcquireDriver(driverName).GenerateQuestionAnswer()
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
func (c *Captcha[K, P, B]) Verify(ctx context.Context, kind K, id, answer string) (bool, error) {
	p, err := c.getParam(kind)
	if err != nil {
		return false, err
	}
	return c.backend.Verify(ctx, &VerifyArgs{
		Key:    p.FormatKey(id),
		Answer: answer,
	})
}

type UnsupportedCaptchaDriver struct{}

func (x UnsupportedCaptchaDriver) Name() string {
	return "Unsupported captcha driver"
}
func (x UnsupportedCaptchaDriver) GenerateQuestionAnswer() (*Challenge, error) {
	return nil, errors.New(x.Name())
}
