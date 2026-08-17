package limit_verified

import (
	"context"
	"errors"
	"math/rand/v2"
	"strconv"
	"time"
)

// ErrSceneParamNotFound is an error that param scene not found.
var ErrSceneParamNotFound = errors.New("limit: the scene's param not found")
var ErrReachMaximumQuota = errors.New("limit: reach the maximum quota")

type SendCodeResult = EvaluateResult

type LimitVerifier interface {
	Name() string
	SendCode(ctx context.Context, target, code string) (*SendCodeResult, error)
	VerifyCode(ctx context.Context, target, code string) (*VerifyResult, error)
}

// LimitVerifiedProvider the provider
type LimitVerifiedProvider interface {
	Name() string
	SendCode(ctx context.Context, target, code string) error
}

type WindowTier struct {
	Window time.Duration // 子窗口时间
	Quota  int           // 子窗口内配额
}

type Param struct {
	Scene           string        // 验证码场景
	Window          time.Duration // 验证码最大滚动窗口时间, 24小时
	Quota           int           // 验证码最大滚动窗口内配额, 30次
	WindowTiers     []WindowTier  // 子窗口限制, 从小到大排列, 如 [{1min,3}, {4h,15}]
	CodeExpires     int           // 验证码有效期, 300秒
	CodeMaxAttempts int           // 验证码最大尝试次数, 3次
}

func NewParam(scene string) *Param {
	return &Param{
		Scene:           scene,
		Window:          time.Hour * 24,
		Quota:           30,
		CodeExpires:     300,
		CodeMaxAttempts: 3,
	}
}

// LimitVerified limit verified code
type LimitVerified[P LimitVerifiedProvider, B LimitVerifiedBackend] struct {
	p         LimitVerifiedProvider // LimitVerifiedProvider send code
	backend   LimitVerifiedBackend  // backend client
	keyPrefix string                // key prefix
	param     *Param
}

// NewLimitVerified  new a limit verified
func NewLimitVerified[P LimitVerifiedProvider, B LimitVerifiedBackend](p P, backend B, param *Param) *LimitVerified[P, B] {
	return &LimitVerified[P, B]{
		p:         p,
		backend:   backend,
		keyPrefix: "limit:verifier:scene:",
		param:     param,
	}
}

// SetKeyPrefix sets the key prefix for the limit verified registry.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (v *LimitVerified[P, B]) SetKeyPrefix(keyPrefix string) *LimitVerified[P, B] {
	v.keyPrefix = keyPrefix
	return v
}

// Name the provider name
func (v *LimitVerified[P, B]) Name() string { return v.p.Name() }

// SendCode send code and backend.
func (v *LimitVerified[P, B]) SendCode(ctx context.Context, target, code string) (*EvaluateResult, error) {
	uniqueId := UniqueId()
	result, err := v.backend.Evaluate(ctx, &EvaluateRequest{
		Key:             v.formatKey(target),
		CodeKey:         v.formatCodeKey(target),
		Window:          v.param.Window,
		Quota:           v.param.Quota,
		WindowTiers:     v.param.WindowTiers,
		CodeExpires:     v.param.CodeExpires,
		CodeMaxAttempts: v.param.CodeMaxAttempts,
		Code:            code,
		UniqueId:        uniqueId,
	})
	if err != nil {
		return nil, err
	}
	if result.Status != EvaluateStatus_Success {
		return result, nil
	}
	// 发送失败, 回滚发送次数
	// 当 provider 达到最大配额时返回 ErrReachMaximumQuota, 此时不需要回滚
	defer func() {
		if err != nil && !errors.Is(err, ErrReachMaximumQuota) {
			_ = v.backend.Rollback(ctx, &RollbackRequest{
				Key:      v.formatKey(target),
				CodeKey:  v.formatCodeKey(target),
				UniqueId: uniqueId,
			})
		}
	}()
	err = v.p.SendCode(ctx, target, code)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// VerifyCode verify code from cache.
func (v *LimitVerified[P, B]) VerifyCode(ctx context.Context, target, code string) (*VerifyResult, error) {
	return v.backend.Verify(ctx, &VerifyRequest{
		Key:     v.formatKey(target),
		CodeKey: v.formatCodeKey(target),
		Code:    code,
	})
}

func (v *LimitVerified[P, B]) formatKey(target string) string {
	return v.keyPrefix + target
}
func (v *LimitVerified[P, B]) formatCodeKey(target string) string {
	return v.keyPrefix + target + ":_code_:" + v.param.Scene
}

// UniqueId 生成一个唯一的id.
func UniqueId() string {
	var buf [20]byte

	b := strconv.AppendUint(buf[:0], uint64(time.Now().UnixNano()), 36)
	b = strconv.AppendUint(b, uint64(rand.Uint32()), 36)
	return string(b)
}
