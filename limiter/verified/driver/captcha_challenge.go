package driver

import (
	"context"

	"github.com/mojocn/base64Captcha"
	"github.com/thinkgos/proc-extra/limiter/verified"
)

var _ verified.ChallengeProvider = (*CaptchaChallenge)(nil)

type CaptchaChallenge struct {
	driver base64Captcha.Driver
}

func NewCaptchaChallenge(d base64Captcha.Driver) *CaptchaChallenge {
	return &CaptchaChallenge{driver: d}
}

func (c *CaptchaChallenge) Name() string { return "base64-captcha" }

func (c *CaptchaChallenge) GenerateChallenge(ctx context.Context) (*verified.Challenge, error) {
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
