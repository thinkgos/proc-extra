package window_limiter

import "context"

type WindowFailureLimiter[S comparable] interface {
	// EvaluateErr see [Evaluate]
	EvaluateErr(ctx context.Context, id string, err error) (*FailureLimiterResult, error)
	// Evaluate  评估本次操作.
	// - 窗口内失败次数超过 MaxFailures, 则 Allow 必定为 false. 并拒绝本次操作.(直接拒绝并提示超过最大限制)
	// - 窗口内失败次数未超过 MaxFailures, 则 Allow 为 true. (走正常流程)
	//   - 若 IsFailure == true, 提示业务错误
	//   - 若 IsFailure == false, 清除所有限制, 走正常流程.
	Evaluate(ctx context.Context, id string, isFailure bool) (*FailureLimiterResult, error)
	// Check 检查下一个操作是否被允许, 不修改任何数据.
	Check(ctx context.Context, id string) (*FailureLimiterResult, error)
	// Lock 锁定 key, 在滑动窗口内将拒绝所有操作.
	Lock(ctx context.Context, id string) (*FailureLimiterResult, error)
	// Reset 清除 key的所有限制, 包括失败记录和锁定.
	Reset(ctx context.Context, id string) error
}

// SlidingWindowFailureLimiterParam 滑动窗口失败限制器参数.
type SlidingWindowFailureLimiterParam struct {
	KeyPrefix   string // key prefix
	Window      int    // sliding window in seconds
	MaxFailures int    // max failures in the sliding window
}

func (p *SlidingWindowFailureLimiterParam) formatKey(id string) string {
	return p.KeyPrefix + id
}
func (p *SlidingWindowFailureLimiterParam) formatLockedKey(id string) string {
	return p.KeyPrefix + id + ":_locked"
}

// SlidingWindowFailureLimiter 滑动窗口失败限制器.
type SlidingWindowFailureLimiter[B SlidingWindowFailureLimiterBackend] struct {
	backend B
	param   *SlidingWindowFailureLimiterParam
}

// NewSlidingWindowFailureLimiter new a SlidingWindowFailureLimiter instance.
func NewSlidingWindowFailureLimiter[B SlidingWindowFailureLimiterBackend](backend B, param *SlidingWindowFailureLimiterParam) *SlidingWindowFailureLimiter[B] {
	return &SlidingWindowFailureLimiter[B]{
		backend: backend,
		param:   param,
	}
}

// EvaluateErr see [Evaluate]
func (l *SlidingWindowFailureLimiter[B]) EvaluateErr(ctx context.Context, id string, err error) (*FailureLimiterResult, error) {
	return l.Evaluate(ctx, id, err != nil)
}

// Evaluate 评估本次操作.
func (l *SlidingWindowFailureLimiter[B]) Evaluate(ctx context.Context, id string, isFailure bool) (*FailureLimiterResult, error) {
	return l.backend.Evaluate(ctx, &FailureLimiterEvaluateRequest{
		Key:         l.param.formatKey(id),
		LockedKey:   l.param.formatLockedKey(id),
		Window:      l.param.Window,
		MaxFailures: l.param.MaxFailures,
		UniqueId:    UniqueId(),
		IsFailure:   isFailure,
	})
}

// Check 检查下一个操作是否被允许, 不修改任何数据.
func (l *SlidingWindowFailureLimiter[B]) Check(ctx context.Context, id string) (*FailureLimiterResult, error) {
	return l.backend.Check(ctx, &FailureLimiterCheckRequest{
		Key:         l.param.formatKey(id),
		LockedKey:   l.param.formatLockedKey(id),
		Window:      l.param.Window,
		MaxFailures: l.param.MaxFailures,
	})
}

// Lock 锁定 key, 在滑动窗口内将拒绝所有操作.
func (l *SlidingWindowFailureLimiter[B]) Lock(ctx context.Context, id string) (*FailureLimiterResult, error) {
	return l.backend.Lock(ctx, &FailureLimiterLockRequest{
		Key:         l.param.formatKey(id),
		LockedKey:   l.param.formatLockedKey(id),
		Window:      l.param.Window,
		MaxFailures: l.param.MaxFailures,
	})
}

// Reset 清除 key的所有限制, 包括失败记录和锁定.
func (l *SlidingWindowFailureLimiter[B]) Reset(ctx context.Context, id string) error {
	return l.backend.Reset(ctx, &FailureLimiterResetRequest{
		Key:       l.param.formatKey(id),
		LockedKey: l.param.formatLockedKey(id),
	})
}
