package window_limiter

import (
	"context"
)

type WindowFailureLimiter[S SceneValuer] interface {
	// EvaluateErr see [Evaluate]
	EvaluateErr(ctx context.Context, scene S, id string, err error) (*FailureLimiterResult, error)
	// Evaluate  评估本次操作.
	// - 窗口内失败次数超过 MaxFailures, 则 Allow 必定为 false. 并拒绝本次操作.(直接拒绝并提示超过最大限制)
	// - 窗口内失败次数未超过 MaxFailures, 则 Allow 为 true. (走正常流程)
	//   - 若 IsFailure == true, 表示业务错误
	//   - 若 IsFailure == false, 清除所有限制, 走正常流程.
	Evaluate(ctx context.Context, scene S, id string, isFailure bool) (*FailureLimiterResult, error)
	// Check 检查下一个操作是否被允许, 不修改任何数据.
	Check(ctx context.Context, scene S, id string) (*FailureLimiterResult, error)
	// Lock 锁定 key, 在滑动窗口内将拒绝所有操作.
	Lock(ctx context.Context, scene S, id string) (*FailureLimiterResult, error)
	// Reset 清除 key的所有限制, 包括失败记录和锁定.
	Reset(ctx context.Context, scene S, id string) error
}

// SlidingWindowFailureLimiter 滑动窗口失败限制器.
type SlidingWindowFailureLimiter[S SceneValuer, B SlidingWindowFailureLimiterBackend] struct {
	backend B
	sps     SceneParamRegistry[S, SlidingWindowLimiterParam]
}

// NewSlidingWindowFailureLimiter new a SlidingWindowFailureLimiter instance.
func NewSlidingWindowFailureLimiter[S SceneValuer, B SlidingWindowFailureLimiterBackend](backend B) *SlidingWindowFailureLimiter[S, B] {
	return &SlidingWindowFailureLimiter[S, B]{
		backend: backend,
		sps: NewSceneParamRegistry[S]("window:failure:limiter:", &SlidingWindowLimiterParam{
			Window:   60,
			MaxLimit: 10,
		}),
	}
}

// SetKeyPrefix sets the key prefix.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (l *SlidingWindowFailureLimiter[S, B]) SetKeyPrefix(keyPrefix string) *SlidingWindowFailureLimiter[S, B] {
	l.sps.SetKeyPrefix(keyPrefix)
	return l
}

// SetGeneralParam sets the general param.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (l *SlidingWindowFailureLimiter[S, B]) SetGeneralParam(p *SlidingWindowLimiterParam) *SlidingWindowFailureLimiter[S, B] {
	l.sps.SetGeneralParam(p)
	return l
}

// SetSceneParam sets the param for a specific scene.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (l *SlidingWindowFailureLimiter[S, B]) SetSceneParam(scene S, param *SlidingWindowLimiterParam) *SlidingWindowFailureLimiter[S, B] {
	l.sps.SetSceneParam(scene, param)
	return l
}

// EvaluateErr see [Evaluate]
func (l *SlidingWindowFailureLimiter[S, B]) EvaluateErr(ctx context.Context, scene S, id string, err error) (*FailureLimiterResult, error) {
	return l.Evaluate(ctx, scene, id, err != nil)
}

// Evaluate 评估本次操作.
func (l *SlidingWindowFailureLimiter[S, B]) Evaluate(ctx context.Context, scene S, id string, isFailure bool) (*FailureLimiterResult, error) {
	p := l.sps.useScene(scene)
	return l.backend.Evaluate(ctx, &FailureLimiterEvaluateRequest{
		Key:         l.sps.formatKey(scene.Value(), id),
		LockedKey:   l.sps.formatLockedKey(scene.Value(), id),
		Window:      p.Window,
		MaxFailures: p.MaxLimit,
		UniqueId:    UniqueId(),
		IsFailure:   isFailure,
	})
}

// Check 检查下一个操作是否被允许, 不修改任何数据.
func (l *SlidingWindowFailureLimiter[S, B]) Check(ctx context.Context, scene S, id string) (*FailureLimiterResult, error) {
	p := l.sps.useScene(scene)
	return l.backend.Check(ctx, &FailureLimiterCheckRequest{
		Key:         l.sps.formatKey(scene.Value(), id),
		LockedKey:   l.sps.formatLockedKey(scene.Value(), id),
		Window:      p.Window,
		MaxFailures: p.MaxLimit,
	})
}

// Lock 锁定 key, 在滑动窗口内将拒绝所有操作.
func (l *SlidingWindowFailureLimiter[S, B]) Lock(ctx context.Context, scene S, id string) (*FailureLimiterResult, error) {
	p := l.sps.useScene(scene)
	return l.backend.Lock(ctx, &FailureLimiterLockRequest{
		Key:         l.sps.formatKey(scene.Value(), id),
		LockedKey:   l.sps.formatLockedKey(scene.Value(), id),
		Window:      p.Window,
		MaxFailures: p.MaxLimit,
	})
}

// Reset 清除 key的所有限制, 包括失败记录和锁定.
func (l *SlidingWindowFailureLimiter[S, B]) Reset(ctx context.Context, scene S, id string) error {
	return l.backend.Reset(ctx, &FailureLimiterResetRequest{
		Key:       l.sps.formatKey(scene.Value(), id),
		LockedKey: l.sps.formatLockedKey(scene.Value(), id),
	})
}
