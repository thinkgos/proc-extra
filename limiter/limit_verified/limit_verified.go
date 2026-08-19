package limit_verified

import (
	"context"
	"errors"
	"math/rand/v2"
	"slices"
	"strconv"
	"time"
)

// ErrReachMaximumQuota is an error that reach the maximum quota.
var ErrReachMaximumQuota = errors.New("limit_verified: reach the maximum quota")

type SendCodeResult = EvaluateResult

type SceneValuer interface {
	comparable
	Value() string
}

type LimitVerifier[S SceneValuer] interface {
	Name() string
	SendCode(ctx context.Context, scene S, target, code string) (*SendCodeResult, error)
	VerifyCode(ctx context.Context, scene S, target, code string) (*VerifyResult, error)
}

// LimitVerified limit verified code
type LimitVerified[S SceneValuer, P LimitVerifiedProvider, B LimitVerifiedBackend] struct {
	p         LimitVerifiedProvider // LimitVerifiedProvider send code
	backend   LimitVerifiedBackend  // backend client
	keyPrefix string                // key prefix
	param     *Param                // wildcard param
	specials  []SpecialParam[S]     // special params for each scene
}

// NewLimitVerified  new a limit verified
func NewLimitVerified[S SceneValuer, P LimitVerifiedProvider, B LimitVerifiedBackend](p P, backend B) *LimitVerified[S, P, B] {
	return &LimitVerified[S, P, B]{
		p:         p,
		backend:   backend,
		keyPrefix: "limit:verifier:",
		param:     NewParam(),
		specials:  make([]SpecialParam[S], 0),
	}
}

// SetKeyPrefix sets the key prefix for the limit verified registry.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (v *LimitVerified[S, P, B]) SetKeyPrefix(keyPrefix string) *LimitVerified[S, P, B] {
	v.keyPrefix = keyPrefix
	return v
}

// SetParam sets the wildcard param for the limit verified registry.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (v *LimitVerified[S, P, B]) SetParam(p *Param) *LimitVerified[S, P, B] {
	v.param = p
	return v
}

// SetSpecialParam sets the param for a specific scene.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (v *LimitVerified[S, P, B]) SetSpecialParam(scene S, param *Param) *LimitVerified[S, P, B] {
	for i := range v.specials {
		if v.specials[i].Scene == scene {
			v.specials[i].Param = param
			return v
		}
	}
	v.specials = append(v.specials, SpecialParam[S]{Scene: scene, Param: param})
	v.specials = slices.Clone(v.specials)
	return v
}
func (v *LimitVerified[S, P, B]) useScene(scene S) *Param {
	for _, s := range v.specials {
		if s.Scene == scene {
			return s.Param
		}
	}
	return v.param
}

// Name the provider name
func (v *LimitVerified[S, P, B]) Name() string { return v.p.Name() }

// SendCode send code and backend.
func (v *LimitVerified[S, P, B]) SendCode(ctx context.Context, scene S, target, code string) (*EvaluateResult, error) {
	p := v.useScene(scene)
	uniqueId := UniqueId()
	key := v.formatKey(target)
	codeKey := v.formatCodeKey(target, scene.Value())
	result, err := v.backend.Evaluate(ctx, &EvaluateRequest{
		Key:             key,
		CodeKey:         codeKey,
		Window:          p.Window,
		Quota:           p.Quota,
		WindowTiers:     p.WindowTiers,
		CodeExpires:     p.CodeExpires,
		CodeMaxAttempts: p.CodeMaxAttempts,
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
				Key:      key,
				CodeKey:  codeKey,
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
	return v.backend.Verify(ctx, &VerifyRequest{
		Key:     v.formatKey(target),
		CodeKey: v.formatCodeKey(target, scene.Value()),
		Code:    code,
	})
}

func (v *LimitVerified[S, P, B]) formatKey(target string) string {
	return v.keyPrefix + target
}
func (v *LimitVerified[S, P, B]) formatCodeKey(target, scene string) string {
	return v.keyPrefix + target + ":_code_:" + scene
}

// UniqueId 生成一个唯一的id.
func UniqueId() string {
	var buf [20]byte

	b := strconv.AppendUint(buf[:0], uint64(time.Now().UnixNano()), 36)
	b = strconv.AppendUint(b, uint64(rand.Uint32()), 36)
	return string(b)
}
