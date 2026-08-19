package verified

import "time"

type SceneValuer interface {
	comparable
	Value() string
}

// Param captcha param
type Param struct {
	KeyExpires  time.Duration // 验证码key的过期时间
	MaxAttempts int           // 验证码最大允许尝试次数
}

func NewParam() *Param {
	return &Param{
		KeyExpires:  time.Minute * 5,
		MaxAttempts: 1,
	}
}

func (p *Param) clone() *Param {
	return &Param{
		KeyExpires:  p.KeyExpires,
		MaxAttempts: p.MaxAttempts,
	}
}
func (p *Param) apply(opts ...Option) *Param {
	for _, f := range opts {
		f(p)
	}
	return p
}

type SceneParam[S SceneValuer] struct {
	scene S
	param *Param
}

// Option param option
type Option func(*Param)

// WithKeyExpires redis存验证码key的过期时间
func WithKeyExpires(t time.Duration) Option {
	return func(p *Param) {
		p.KeyExpires = t
	}
}

// WithMaxAttempts 设置最大允许尝试次数
func WithMaxAttempts(attempts int) Option {
	return func(p *Param) {
		if attempts > 0 {
			p.MaxAttempts = attempts
		}
	}
}
