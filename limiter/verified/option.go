package verified

import (
	"errors"
	"time"
)

// ErrParamKindNotFound is an error that param kind not found.
var ErrParamKindNotFound = errors.New("limit: param kind not found")

// Param captcha param
type Param struct {
	KeyPrefix   string        // 验证码key前缀
	KeyExpires  time.Duration // 验证码key的过期时间
	MaxAttempts int           // 验证码最大允许尝试次数
}

func (p *Param) FormatKey(id string) string {
	return p.KeyPrefix + id
}

func (p *Param) clone() *Param {
	return &Param{
		KeyPrefix:   p.KeyPrefix,
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
		if attempts > 1 {
			p.MaxAttempts = attempts
		}
	}
}
