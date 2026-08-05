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

type LimitVerifier[S comparable] interface {
	Name() string
	SendCode(ctx context.Context, scene S, target, code string) (*SendCodeResult, error)
	VerifyCode(ctx context.Context, scene S, target, code string) (*VerifyResult, error)
}

// LimitVerifiedProvider the provider
type LimitVerifiedProvider interface {
	Name() string
	SendCode(ctx context.Context, target, code string) error
}

type Param struct {
	KeyPrefix       string        // 验证码key的前缀.
	Window          time.Duration // 验证码滚动窗口时间, 24小时
	Quota           int           // 验证码滚动窗口内配额, 30次
	ResendInterval  int           // 验证码重发间隔时间, 60秒
	CodeExpires     int           // 验证码有效期, 300秒
	CodeMaxAttempts int           // 验证码最大尝试次数, 3次
}

func NewParam(keyPrefix string) *Param {
	return &Param{
		KeyPrefix:       keyPrefix,
		Window:          time.Hour * 24,
		Quota:           30,
		ResendInterval:  60,
		CodeExpires:     300,
		CodeMaxAttempts: 3,
	}
}

func (p *Param) FormatKey(id string) string { return p.KeyPrefix + id }

// LimitVerified limit verified code
type LimitVerified[S comparable, P LimitVerifiedProvider, B LimitVerifiedBackend] struct {
	p       LimitVerifiedProvider // LimitVerifiedProvider send code
	backend LimitVerifiedBackend  // backend client
	params  map[S]*Param
}

// NewLimitVerified  new a limit verified
func NewLimitVerified[S comparable, P LimitVerifiedProvider, B LimitVerifiedBackend](p P, backend B) *LimitVerified[S, P, B] {
	v := &LimitVerified[S, P, B]{
		p:       p,
		backend: backend,
		params:  make(map[S]*Param),
	}
	return v
}

func (v *LimitVerified[S, P, B]) SetParams(params map[S]*Param) *LimitVerified[S, P, B] {
	v.params = params
	return v
}
func (v *LimitVerified[S, P, B]) getParam(scene S) (*Param, error) {
	param, ok := v.params[scene]
	if !ok {
		return nil, ErrSceneParamNotFound
	}
	return param, nil
}

// Name the provider name
func (v *LimitVerified[S, P, B]) Name() string { return v.p.Name() }

// SendCode send code and backend.
func (v *LimitVerified[S, P, B]) SendCode(ctx context.Context, scene S, target, code string) (*EvaluateResult, error) {
	p, err := v.getParam(scene)
	if err != nil {
		return nil, err
	}
	uniqueId := UniqueId()
	result, err := v.backend.Evaluate(ctx, &EvaluateRequest{
		Key:             p.FormatKey(target),
		Window:          p.Window,
		Quota:           p.Quota,
		ResendInterval:  p.ResendInterval,
		CodeExpires:     p.CodeExpires,
		CodeMaxAttempts: p.CodeMaxAttempts,
		Code:            code,
		UniqueId:        uniqueId,
	})
	if err != nil {
		return nil, err
	}
	if result.Code != EvaluateCode_Success {
		return result, nil
	}
	// 发送失败, 回滚发送次数
	defer func() {
		if err != nil && !errors.Is(err, ErrReachMaximumQuota) {
			_ = v.backend.Rollback(context.Background(), &RollbackRequest{
				Key:      p.FormatKey(target),
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
func (v *LimitVerified[S, P, B]) VerifyCode(ctx context.Context, scene S, target, code string) (*VerifyResult, error) {
	p, err := v.getParam(scene)
	if err != nil {
		return nil, err
	}
	return v.backend.Verify(ctx, &VerifyRequest{
		Key:  p.FormatKey(target),
		Code: code,
	})
}

// UniqueId 生成一个唯一的id.
func UniqueId() string {
	var buf [20]byte

	b := strconv.AppendUint(buf[:0], uint64(time.Now().UnixNano()), 36)
	b = strconv.AppendUint(b, uint64(rand.Uint32()), 36)
	return string(b)
}
