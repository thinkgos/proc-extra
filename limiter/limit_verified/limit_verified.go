package limit_verified

import (
	"context"
	"errors"
	"math/rand/v2"
	"strconv"
	"time"
)

// ErrParamKindNotFound is an error that param kind not found.
var ErrParamKindNotFound = errors.New("limit: param kind not found")
var ErrReachMaximumQuota = errors.New("limit: reach the maximum quota")

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
type LimitVerified[K comparable, P LimitVerifiedProvider, B LimitVerifiedBackend] struct {
	p       LimitVerifiedProvider // LimitVerifiedProvider send code
	backend LimitVerifiedBackend  // backend client
	params  map[K]*Param
}

// NewLimitVerified  new a limit verified
func NewLimitVerified[K comparable, P LimitVerifiedProvider, B LimitVerifiedBackend](p P, backend B) *LimitVerified[K, P, B] {
	v := &LimitVerified[K, P, B]{
		p:       p,
		backend: backend,
		params:  make(map[K]*Param),
	}
	return v
}

func (v *LimitVerified[K, P, B]) SetParams(params map[K]*Param) *LimitVerified[K, P, B] {
	v.params = params
	return v
}
func (v *LimitVerified[K, P, B]) getParam(kind K) (*Param, error) {
	param, ok := v.params[kind]
	if !ok {
		return nil, ErrParamKindNotFound
	}
	return param, nil
}

// Name the provider name
func (v *LimitVerified[K, P, B]) Name() string { return v.p.Name() }

// SendCode send code and backend.
func (v *LimitVerified[K, P, B]) SendCode(ctx context.Context, kind K, target, code string) (*EvaluateResult, error) {
	p, err := v.getParam(kind)
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
func (v *LimitVerified[K, P, B]) VerifyCode(ctx context.Context, kind K, target, code string) (*VerifyResult, error) {
	p, err := v.getParam(kind)
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
