package limit_verified

import (
	"context"
	"time"
)

type EvaluateStatus int

const (
	// EvaluateStatus_Success 发送成功
	EvaluateStatus_Success EvaluateStatus = iota
	// EvaluateStatus_OverQuota 发送超过配额
	EvaluateStatus_OverQuota
	// EvaluateStatus_TooFrequently 发送过于频繁
	EvaluateStatus_TooFrequently
)

type VerifyStatus int

const (
	// VerifyStatus_Success 验证成功
	VerifyStatus_Success VerifyStatus = iota
	// VerifyStatus_Failure 验证失败
	VerifyStatus_Failure
	// VerifyStatus_Expired 验证码已失效
	VerifyStatus_Expired
)

// EvaluateRequest store arguments
type EvaluateRequest struct {
	Key             string        // 目标键
	CodeKey         string        // 验证码键
	Window          time.Duration // 验证码最大滚动窗口时间, 24小时
	Quota           int           // 验证码最大滚动窗口内配额, 30次
	WindowTiers     []WindowTier  // 子窗口限制
	CodeExpires     int           // 验证码有效期, 180秒
	CodeMaxAttempts int           // 验证码最大允许尝试次数, 3次
	Code            string        // 验证码
	UniqueId        string        // 唯一id
}
type EvaluateResult struct {
	Status EvaluateStatus
}

type RollbackRequest struct {
	Key      string // 目标键
	CodeKey  string // 验证码键
	UniqueId string // 唯一id
}

type VerifyRequest struct {
	Key     string // 验证码键
	CodeKey string // 验证码键
	Code    string // 验证码
}
type VerifyResult struct {
	Status VerifyStatus
}

type LimitVerifiedBackend interface {
	// Evaluate
	Evaluate(context.Context, *EvaluateRequest) (*EvaluateResult, error)
	// Rollback when evaluate success but send code failed.
	Rollback(context.Context, *RollbackRequest) error
	// Verify code
	Verify(context.Context, *VerifyRequest) (*VerifyResult, error)
}
