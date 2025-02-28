package captcha

import (
	"github.com/mojocn/base64Captcha"
	"github.com/things-go/limiter/verified"
)

var _ verified.CaptchaDriver = (*Captcha)(nil)

type Captcha struct {
	d base64Captcha.Driver
}

func New(d base64Captcha.Driver) *Captcha {
	return &Captcha{
		d,
	}
}

func (c *Captcha) Name() string { return "base64Captcha" }

func (c *Captcha) GenerateQuestionAnswer() (*verified.QuestionAnswer, error) {
	id, q, a := c.d.GenerateIdQuestionAnswer()
	it, err := c.d.DrawCaptcha(q)
	if err != nil {
		return nil, err
	}
	return &verified.QuestionAnswer{
		Id:       id,
		Question: it.EncodeB64string(),
		Answer:   a,
	}, nil
}
