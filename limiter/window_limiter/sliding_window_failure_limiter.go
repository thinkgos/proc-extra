package window_limiter

import "context"

// SlidingWindowFailureLimiterParam 滑动窗口失败限制器参数.
type SlidingWindowFailureLimiterParam struct {
	KeyPrefix   string // key prefix
	Window      int    // sliding window in seconds
	MaxFailures int    // max failures in the sliding window
}

func (p *SlidingWindowFailureLimiterParam) FormatKey(id string) string { return p.KeyPrefix + id }

// SlidingWindowFailureLimiter 滑动窗口失败限制器.
type SlidingWindowFailureLimiter[K comparable, B SlidingWindowFailureLimiterBackend] struct {
	backend B
	params  map[K]*SlidingWindowFailureLimiterParam
}

// NewSlidingWindowFailureLimiter 创建新的 SlidingWindowFailureLimiter 实例.
func NewSlidingWindowFailureLimiter[K comparable, B SlidingWindowFailureLimiterBackend](backend B) *SlidingWindowFailureLimiter[K, B] {
	return &SlidingWindowFailureLimiter[K, B]{
		backend: backend,
		params:  make(map[K]*SlidingWindowFailureLimiterParam),
	}
}

// SetParams 设置参数.
func (p *SlidingWindowFailureLimiter[K, B]) SetParams(params map[K]*SlidingWindowFailureLimiterParam) *SlidingWindowFailureLimiter[K, B] {
	p.params = params
	return p
}

func (p *SlidingWindowFailureLimiter[K, B]) getParam(kind K) (*SlidingWindowFailureLimiterParam, error) {
	param, ok := p.params[kind]
	if !ok {
		return nil, ErrParamKindNotFound
	}
	return param, nil
}

// EvaluateErr see [Evaluate]
func (p *SlidingWindowFailureLimiter[K, B]) EvaluateErr(ctx context.Context, kind K, id string, err error) (*FailureLimiterResult, error) {
	return p.Evaluate(ctx, kind, id, err != nil)
}

// Evaluate  评估本次操作.
// - 窗口内失败次数超过 MaxFailures, 则 Allow 必定为 false. 并拒绝本次操作.(直接拒绝并提示超过最大限制)
// - 窗口内失败次数未超过 MaxFailures, 则 Allow 为 true. (走正常流程)
//   - 若 IsFailure == true, 提示业务错误
//   - 若 IsFailure == false, 清除所有限制, 走正常流程.
func (p *SlidingWindowFailureLimiter[K, B]) Evaluate(ctx context.Context, kind K, id string, isFailure bool) (*FailureLimiterResult, error) {
	pm, err := p.getParam(kind)
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
func (p *SlidingWindowFailureLimiter[K, B]) Check(ctx context.Context, kind K, id string) (*FailureLimiterResult, error) {
	pm, err := p.getParam(kind)
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
func (p *SlidingWindowFailureLimiter[K, B]) Lock(ctx context.Context, kind K, id string) (*FailureLimiterResult, error) {
	pm, err := p.getParam(kind)
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
func (p *SlidingWindowFailureLimiter[K, B]) Reset(ctx context.Context, kind K, id string) error {
	pm, err := p.getParam(kind)
	if err != nil {
		return err
	}
	return p.backend.Reset(ctx, pm.FormatKey(id))
}
