package limit_verified

import (
	"context"
	"slices"
	"time"
)

type WindowTier struct {
	Window time.Duration // 子窗口时间
	Quota  int           // 子窗口内配额
}

type Param struct {
	Window          time.Duration // 验证码最大滚动窗口时间, 24小时
	Quota           int           // 验证码最大滚动窗口内配额, 30次
	WindowTiers     []WindowTier  // 子窗口限制, 从小到大排列, 如 [{1min,1}, {4h,5}]
	CodeExpires     int           // 验证码有效期, 300秒
	CodeMaxAttempts int           // 验证码最大尝试次数, 3次
}

func NewParam() *Param {
	return &Param{
		Window:          time.Hour * 24,
		Quota:           30,
		WindowTiers:     []WindowTier{{time.Second * 60, 1}},
		CodeExpires:     300,
		CodeMaxAttempts: 3,
	}
}

func (p *Param) Clone() *Param {
	return &Param{
		Window:          p.Window,
		Quota:           p.Quota,
		CodeExpires:     p.CodeExpires,
		CodeMaxAttempts: p.CodeMaxAttempts,
		WindowTiers:     slices.Clone(p.WindowTiers),
	}
}

type SpecialParam [S SceneValuer]struct {
	Scene S
	Param *Param
}

// LimitVerifiedProvider the provider
type LimitVerifiedProvider interface {
	Name() string
	SendCode(ctx context.Context, target, code string) error
}

var _ LimitVerifiedProvider = DummyDriver{}

type DummyDriver struct{}

func (DummyDriver) Name() string                                            { return "limit-verified-dummy-provider" }
func (DummyDriver) SendCode(ctx context.Context, target, code string) error { return nil }
