package window_limiter

import "context"

type WindowFailureLimiterRegistry[S comparable] interface {
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

// SlidingWindowFailureLimiterRegistry 滑动窗口失败限制器.
type SlidingWindowFailureLimiterRegistry[S comparable, B SlidingWindowFailureLimiterBackend] struct {
	backend B
	params  map[S]*SlidingWindowFailureLimiterParam
}

// NewSlidingWindowFailureLimiterRegistry 创建新的 SlidingWindowFailureLimiter 实例.
func NewSlidingWindowFailureLimiterRegistry[S comparable, B SlidingWindowFailureLimiterBackend](backend B) *SlidingWindowFailureLimiterRegistry[S, B] {
	return &SlidingWindowFailureLimiterRegistry[S, B]{
		backend: backend,
		params:  make(map[S]*SlidingWindowFailureLimiterParam),
	}
}

// SetParams 设置参数.
func (r *SlidingWindowFailureLimiterRegistry[S, B]) SetParams(params map[S]*SlidingWindowFailureLimiterParam) *SlidingWindowFailureLimiterRegistry[S, B] {
	r.params = params
	return r
}

func (r *SlidingWindowFailureLimiterRegistry[S, B]) SetParam(scene S, param *SlidingWindowFailureLimiterParam) *SlidingWindowFailureLimiterRegistry[S, B] {
	r.params[scene] = param
	return r
}

func (r *SlidingWindowFailureLimiterRegistry[S, B]) getParam(scene S) (*SlidingWindowFailureLimiterParam, error) {
	param, ok := r.params[scene]
	if !ok {
		return nil, ErrSceneParamNotFound
	}
	return param, nil
}

// EvaluateErr see [Evaluate]
func (r *SlidingWindowFailureLimiterRegistry[S, B]) EvaluateErr(ctx context.Context, scene S, id string, err error) (*FailureLimiterResult, error) {
	return r.Evaluate(ctx, scene, id, err != nil)
}

// Evaluate 评估本次操作.
func (r *SlidingWindowFailureLimiterRegistry[S, B]) Evaluate(ctx context.Context, scene S, id string, isFailure bool) (*FailureLimiterResult, error) {
	pm, err := r.getParam(scene)
	if err != nil {
		return nil, err
	}
	return NewSlidingWindowFailureLimiter(r.backend, pm).Evaluate(ctx, id, isFailure)
}

// Check 检查下一个操作是否被允许, 不修改任何数据.
func (r *SlidingWindowFailureLimiterRegistry[S, B]) Check(ctx context.Context, scene S, id string) (*FailureLimiterResult, error) {
	pm, err := r.getParam(scene)
	if err != nil {
		return nil, err
	}
	return NewSlidingWindowFailureLimiter(r.backend, pm).Check(ctx, id)
}

// Lock 锁定 key, 在滑动窗口内将拒绝所有操作.
func (r *SlidingWindowFailureLimiterRegistry[S, B]) Lock(ctx context.Context, scene S, id string) (*FailureLimiterResult, error) {
	pm, err := r.getParam(scene)
	if err != nil {
		return nil, err
	}
	return NewSlidingWindowFailureLimiter(r.backend, pm).Lock(ctx, id)
}

// Reset 清除 key的所有限制, 包括失败记录和锁定.
func (r *SlidingWindowFailureLimiterRegistry[S, B]) Reset(ctx context.Context, scene S, id string) error {
	pm, err := r.getParam(scene)
	if err != nil {
		return err
	}
	return NewSlidingWindowFailureLimiter(r.backend, pm).Reset(ctx, id)
}
