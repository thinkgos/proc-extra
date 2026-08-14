package limit_verified

import (
	"context"
	"time"
)

type EvaluateCode int

const (
	// EvaluateCode_Success 发送成功
	EvaluateCode_Success EvaluateCode = iota
	// EvaluateCode_OverQuota 发送超过配额
	EvaluateCode_OverQuota
	// EvaluateCode_TooFrequently 发送过于频繁
	EvaluateCode_TooFrequently
)

type VerifyCode int

const (
	// VerifyCode_Success 验证成功
	VerifyCode_Success VerifyCode = iota
	// VerifyCode_Failure 验证失败
	VerifyCode_Failure
	// VerifyCode_Expired 验证码已失效
	VerifyCode_Expired
)

// EvaluateRequest store arguments
type EvaluateRequest struct {
	KeyPrefix       string        // 验证码键前缀
	Target          string        // 目标
	Scene           string        // 场景
	Window          time.Duration // 验证码滚动窗口时间, 24小时
	Quota           int           // 验证码滚动窗口内配额, 30次
	ResendInterval  int           // 验证码重发间隔时间, 60秒
	CodeExpires     int           // 验证码有效期, 180秒
	CodeMaxAttempts int           // 验证码最大允许尝试次数, 3次
	Code            string        // 验证码
	UniqueId        string        // 唯一id
}
type EvaluateResult struct {
	Code EvaluateCode
}

type RollbackRequest struct {
	KeyPrefix string // 验证码键前缀
	Target    string // 目标
	Scene     string // 场景
	UniqueId  string // 唯一id
}

type VerifyRequest struct {
	KeyPrefix string // 验证码键前缀
	Target    string // 目标
	Scene     string // 场景
	Code      string // 验证码
}
type VerifyResult struct {
	Code VerifyCode
}

type LimitVerifiedBackend interface {
	// Evaluate
	Evaluate(context.Context, *EvaluateRequest) (*EvaluateResult, error)
	// Rollback when evaluate success but send code failed.
	Rollback(context.Context, *RollbackRequest) error
	// Verify code
	Verify(context.Context, *VerifyRequest) (*VerifyResult, error)
}
