package window_limiter

import (
	"context"
)

type SlidingWindowLimiterParam struct {
	KeyPrefix string // store key prefix
	Window    int    // sliding window in seconds
	MaxLimit  int    // max limit requests in the sliding window
}

func (p *SlidingWindowLimiterParam) FormatKey(id string) string { return p.KeyPrefix + id }

type SlidingWindowLimiter[K comparable, B SlidingWindowLimiterBackend] struct {
	backend B
	params  map[K]*SlidingWindowLimiterParam
}

func NewSlidingWindowLimiter[K comparable, B SlidingWindowLimiterBackend](backend B) *SlidingWindowLimiter[K, B] {
	return &SlidingWindowLimiter[K, B]{
		backend: backend,
		params:  make(map[K]*SlidingWindowLimiterParam),
	}
}

func (p *SlidingWindowLimiter[K, B]) SetParams(params map[K]*SlidingWindowLimiterParam) *SlidingWindowLimiter[K, B] {
	p.params = params
	return p
}

func (p *SlidingWindowLimiter[K, B]) getParam(kind K) (*SlidingWindowLimiterParam, error) {
	param, ok := p.params[kind]
	if !ok {
		return nil, ErrParamKindNotFound
	}
	return param, nil
}

// Take 尝试获取一个请求的配额单位.
// 如果有可用配额, 则请求被允许, 并且增加一次配额消费.
// 如果没有配额, 则请求被拒绝.
func (p *SlidingWindowLimiter[K, B]) Take(ctx context.Context, kind K, id string) (*LimiterResult, error) {
	pm, err := p.getParam(kind)
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
func (p *SlidingWindowLimiter[K, B]) Check(ctx context.Context, kind K, id string) (*LimiterResult, error) {
	pm, err := p.getParam(kind)
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
func (p *SlidingWindowLimiter[K, B]) Lock(ctx context.Context, kind K, id string) (*LimiterResult, error) {
	pm, err := p.getParam(kind)
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
func (p *SlidingWindowLimiter[K, B]) Reset(ctx context.Context, kind K, id string) error {
	pm, err := p.getParam(kind)
	if err != nil {
		return err
	}
	return p.backend.Reset(ctx, pm.FormatKey(id))
}
