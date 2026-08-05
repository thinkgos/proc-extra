package window_limiter

import (
	"context"
)

type WindowLimiter[S comparable] interface {
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

type SlidingWindowLimiterParam struct {
	KeyPrefix string // store key prefix
	Window    int    // sliding window in seconds
	MaxLimit  int    // max limit requests in the sliding window
}

func (p *SlidingWindowLimiterParam) FormatKey(id string) string { return p.KeyPrefix + id }

type SlidingWindowLimiter[S comparable, B SlidingWindowLimiterBackend] struct {
	backend B
	params  map[S]*SlidingWindowLimiterParam
}

func NewSlidingWindowLimiter[S comparable, B SlidingWindowLimiterBackend](backend B) *SlidingWindowLimiter[S, B] {
	return &SlidingWindowLimiter[S, B]{
		backend: backend,
		params:  make(map[S]*SlidingWindowLimiterParam),
	}
}

func (p *SlidingWindowLimiter[S, B]) SetParams(params map[S]*SlidingWindowLimiterParam) *SlidingWindowLimiter[S, B] {
	p.params = params
	return p
}

func (p *SlidingWindowLimiter[S, B]) getParam(scene S) (*SlidingWindowLimiterParam, error) {
	param, ok := p.params[scene]
	if !ok {
		return nil, ErrSceneParamNotFound
	}
	return param, nil
}

// Take 尝试获取一个请求的配额单位.
// 如果有可用配额, 则请求被允许, 并且增加一次配额消费.
// 如果没有配额, 则请求被拒绝.
func (p *SlidingWindowLimiter[S, B]) Take(ctx context.Context, scene S, id string) (*LimiterResult, error) {
	pm, err := p.getParam(scene)
	if err != nil {
		return nil, err
	}
	return p.backend.Take(ctx, &LimiterTakeRequest{
		Key:      pm.FormatKey(id),
		Window:   pm.Window,
		MaxLimit: pm.MaxLimit,
		UniqueId: UniqueId(),
	})
}

// Check 检查下一个请求是否被允许, 不修改任何数据.
func (p *SlidingWindowLimiter[S, B]) Check(ctx context.Context, scene S, id string) (*LimiterResult, error) {
	pm, err := p.getParam(scene)
	if err != nil {
		return nil, err
	}
	return p.backend.Check(ctx, &LimiterCheckRequest{
		Key:      pm.FormatKey(id),
		Window:   pm.Window,
		MaxLimit: pm.MaxLimit,
	})
}

// Lock 锁定 key, 在滑动窗口内将拒绝所有请求.
func (p *SlidingWindowLimiter[S, B]) Lock(ctx context.Context, scene S, id string) (*LimiterResult, error) {
	pm, err := p.getParam(scene)
	if err != nil {
		return nil, err
	}
	return p.backend.Lock(ctx, &LimiterLockRequest{
		Key:      pm.FormatKey(id),
		Window:   pm.Window,
		MaxLimit: pm.MaxLimit,
	})
}

// Reset 清除 key的所有限制.
func (p *SlidingWindowLimiter[S, B]) Reset(ctx context.Context, scene S, id string) error {
	pm, err := p.getParam(scene)
	if err != nil {
		return err
	}
	return p.backend.Reset(ctx, pm.FormatKey(id))
}
