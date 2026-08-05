package window_limiter

import "context"

type WindowFailureLimiter[S comparable] interface {
	// EvaluateErr see [Evaluate]
	EvaluateErr(ctx context.Context, scene S, id string, err error) (*FailureLimiterResult, error)
	// Evaluate  评估本次操作.
	// - 窗口内失败次数超过 MaxFailures, 则 Allow 必定为 false. 并拒绝本次操作.(直接拒绝并提示超过最大限制)
	// - 窗口内失败次数未超过 MaxFailures, 则 Allow 为 true. (走正常流程)
	//   - 若 IsFailure == true, 提示业务错误
	//   - 若 IsFailure == false, 清除所有限制, 走正常流程.
	Evaluate(ctx context.Context, scene S, id string, isFailure bool) (*FailureLimiterResult, error)
	// Check 检查下一个操作是否被允许, 不修改任何数据.
	Check(ctx context.Context, scene S, id string) (*FailureLimiterResult, error)
	// Lock 锁定 key, 在滑动窗口内将拒绝所有操作.
	Lock(ctx context.Context, scene S, id string) (*FailureLimiterResult, error)
	// Reset 清除 key的所有限制, 包括失败记录和锁定.
	Reset(ctx context.Context, scene S, id string) error
}

// SlidingWindowFailureLimiterParam 滑动窗口失败限制器参数.
type SlidingWindowFailureLimiterParam struct {
	KeyPrefix   string // key prefix
	Window      int    // sliding window in seconds
	MaxFailures int    // max failures in the sliding window
}

func (p *SlidingWindowFailureLimiterParam) FormatKey(id string) string { return p.KeyPrefix + id }

// SlidingWindowFailureLimiter 滑动窗口失败限制器.
type SlidingWindowFailureLimiter[S comparable, B SlidingWindowFailureLimiterBackend] struct {
	backend B
	params  map[S]*SlidingWindowFailureLimiterParam
}

// NewSlidingWindowFailureLimiter 创建新的 SlidingWindowFailureLimiter 实例.
func NewSlidingWindowFailureLimiter[S comparable, B SlidingWindowFailureLimiterBackend](backend B) *SlidingWindowFailureLimiter[S, B] {
	return &SlidingWindowFailureLimiter[S, B]{
		backend: backend,
		params:  make(map[S]*SlidingWindowFailureLimiterParam),
	}
}

// SetParams 设置参数.
func (p *SlidingWindowFailureLimiter[S, B]) SetParams(params map[S]*SlidingWindowFailureLimiterParam) *SlidingWindowFailureLimiter[S, B] {
	p.params = params
	return p
}

func (p *SlidingWindowFailureLimiter[S, B]) getParam(scene S) (*SlidingWindowFailureLimiterParam, error) {
	param, ok := p.params[scene]
	if !ok {
		return nil, ErrSceneParamNotFound
	}
	return param, nil
}

// EvaluateErr see [Evaluate]
func (p *SlidingWindowFailureLimiter[S, B]) EvaluateErr(ctx context.Context, scene S, id string, err error) (*FailureLimiterResult, error) {
	return p.Evaluate(ctx, scene, id, err != nil)
}

// Evaluate 评估本次操作.
func (p *SlidingWindowFailureLimiter[S, B]) Evaluate(ctx context.Context, scene S, id string, isFailure bool) (*FailureLimiterResult, error) {
	pm, err := p.getParam(scene)
	if err != nil {
		return nil, err
	}
	return p.backend.Evaluate(ctx, &FailureLimiterEvaluateRequest{
		Key:         pm.FormatKey(id),
		Window:      pm.Window,
		MaxFailures: pm.MaxFailures,
		UniqueId:    UniqueId(),
		IsFailure:   isFailure,
	})
}

// Check 检查下一个操作是否被允许, 不修改任何数据.
func (p *SlidingWindowFailureLimiter[S, B]) Check(ctx context.Context, scene S, id string) (*FailureLimiterResult, error) {
	pm, err := p.getParam(scene)
	if err != nil {
		return nil, err
	}
	return p.backend.Check(ctx, &FailureLimiterCheckRequest{
		Key:         pm.FormatKey(id),
		Window:      pm.Window,
		MaxFailures: pm.MaxFailures,
	})
}

// Lock 锁定 key, 在滑动窗口内将拒绝所有操作.
func (p *SlidingWindowFailureLimiter[S, B]) Lock(ctx context.Context, scene S, id string) (*FailureLimiterResult, error) {
	pm, err := p.getParam(scene)
	if err != nil {
		return nil, err
	}
	return p.backend.Lock(ctx, &FailureLimiterLockRequest{
		Key:         pm.FormatKey(id),
		Window:      pm.Window,
		MaxFailures: pm.MaxFailures,
	})
}

// Reset 清除 key的所有限制, 包括失败记录和锁定.
func (p *SlidingWindowFailureLimiter[S, B]) Reset(ctx context.Context, scene S, id string) error {
	pm, err := p.getParam(scene)
	if err != nil {
		return err
	}
	return p.backend.Reset(ctx, pm.FormatKey(id))
}
