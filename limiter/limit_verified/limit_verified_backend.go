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
	Status EvaluateStatus
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
