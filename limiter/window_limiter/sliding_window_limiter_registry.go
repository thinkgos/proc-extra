package window_limiter

import (
	"context"
	"maps"
)

type WindowLimiterRegistry[S comparable] interface {
	// Take 尝试获取一个请求的配额单位.
	// 如果有可用配额, 则请求被允许, 并且增加一次配额消费.
	// 如果没有配额, 则请求被拒绝.
	Take(ctx context.Context, scene S, id string) (*LimiterResult, error)
	// Check 检查下一个请求是否被允许, 不修改任何数据.
	Check(ctx context.Context, scene S, id string) (*LimiterResult, error)
	// Lock 锁定 key, 在滑动窗口内将拒绝所有请求.
	Lock(ctx context.Context, scene S, id string) (*LimiterResult, error)
	// Reset 清除 key的所有限制.
	Reset(ctx context.Context, scene S, id string) error
}

type SlidingWindowLimiterRegistry[S comparable, B SlidingWindowLimiterBackend] struct {
	backend B
	params  map[S]*SlidingWindowLimiterParam
}

func NewSlidingWindowLimiterRegistry[S comparable, B SlidingWindowLimiterBackend](backend B) *SlidingWindowLimiterRegistry[S, B] {
	return &SlidingWindowLimiterRegistry[S, B]{
		backend: backend,
		params:  make(map[S]*SlidingWindowLimiterParam),
	}
}

// RegistersParams sets all scene params at once.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (r *SlidingWindowLimiterRegistry[S, B]) RegistersParams(params map[S]*SlidingWindowLimiterParam) *SlidingWindowLimiterRegistry[S, B] {
	maps.Copy(r.params, params)
	return r
}

// RegisterParam sets the param for a specific scene.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (r *SlidingWindowLimiterRegistry[S, B]) RegisterParam(scene S, param *SlidingWindowLimiterParam) *SlidingWindowLimiterRegistry[S, B] {
	r.params[scene] = param
	return r
}

func (r *SlidingWindowLimiterRegistry[S, B]) getParam(scene S) (*SlidingWindowLimiterParam, error) {
	param, ok := r.params[scene]
	if !ok {
		return nil, ErrSceneParamNotFound
	}
	return param, nil
}

// Take 尝试获取一个请求的配额单位.
// 如果有可用配额, 则请求被允许, 并且增加一次配额消费.
// 如果没有配额, 则请求被拒绝.
func (r *SlidingWindowLimiterRegistry[S, B]) Take(ctx context.Context, scene S, id string) (*LimiterResult, error) {
	pm, err := r.getParam(scene)
	if err != nil {
		return nil, err
	}
	return NewSlidingWindowLimiter(r.backend, pm).Take(ctx, id)
}

// Check 检查下一个请求是否被允许, 不修改任何数据.
func (r *SlidingWindowLimiterRegistry[S, B]) Check(ctx context.Context, scene S, id string) (*LimiterResult, error) {
	pm, err := r.getParam(scene)
	if err != nil {
		return nil, err
	}
	return NewSlidingWindowLimiter(r.backend, pm).Check(ctx, id)
}

// Lock 锁定 key, 在滑动窗口内将拒绝所有请求.
func (r *SlidingWindowLimiterRegistry[S, B]) Lock(ctx context.Context, scene S, id string) (*LimiterResult, error) {
	pm, err := r.getParam(scene)
	if err != nil {
		return nil, err
	}
	return NewSlidingWindowLimiter(r.backend, pm).Lock(ctx, id)
}

// Reset 清除 key的所有限制.
func (r *SlidingWindowLimiterRegistry[S, B]) Reset(ctx context.Context, scene S, id string) error {
	pm, err := r.getParam(scene)
	if err != nil {
		return err
	}
	return NewSlidingWindowLimiter(r.backend, pm).Reset(ctx, id)
}
