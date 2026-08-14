package driver

import (
	"context"

	"github.com/mojocn/base64Captcha"
	"github.com/thinkgos/proc-extra/limiter/verified"
)

var _ verified.ChallengeProvider = (*Captcha)(nil)

type Captcha struct {
	driver base64Captcha.Driver
}

func NewCaptcha(d base64Captcha.Driver) *Captcha {
	return &Captcha{driver: d}
}

func (c *Captcha) Name() string { return "base64-captcha" }

func (c *Captcha) GenerateChallenge(ctx context.Context) (*verified.Challenge, error) {
	id, q, a := c.driver.GenerateIdQuestionAnswer()
	it, err := c.driver.DrawCaptcha(q)
	if err != nil {
		return nil, err
	}
	return &verified.Challenge{
		Id:       id,
		Question: it.EncodeB64string(),
		Answer:   a,
	}, nil
}
