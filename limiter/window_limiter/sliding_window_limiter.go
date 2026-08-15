package window_limiter

import "context"

type WindowLimiter[S comparable] interface {
	// Take 尝试获取一个请求的配额单位.
	// 如果有可用配额, 则请求被允许, 并且增加一次配额消费.
	// 如果没有配额, 则请求被拒绝.
	Take(ctx context.Context, id string) (*LimiterResult, error)
	// Check 检查下一个请求是否被允许, 不修改任何数据.
	Check(ctx context.Context, id string) (*LimiterResult, error)
	// Lock 锁定 key, 在滑动窗口内将拒绝所有请求.
	Lock(ctx context.Context, id string) (*LimiterResult, error)
	// Reset 清除 key的所有限制.
	Reset(ctx context.Context, id string) error
}

type SlidingWindowLimiterParam struct {
	KeyPrefix string // store key prefix
	Window    int    // sliding window in seconds
	MaxLimit  int    // max limit requests in the sliding window
}

func (p *SlidingWindowLimiterParam) formatKey(id string) string { return p.KeyPrefix + id }

type SlidingWindowLimiter[B SlidingWindowLimiterBackend] struct {
	backend B
	param   *SlidingWindowLimiterParam
}

func NewSlidingWindowLimiter[B SlidingWindowLimiterBackend](backend B, param *SlidingWindowLimiterParam) SlidingWindowLimiter[B] {
	return SlidingWindowLimiter[B]{
		backend: backend,
		param:   param,
	}
}

// Take 尝试获取一个请求的配额单位.
// 如果有可用配额, 则请求被允许, 并且增加一次配额消费.
// 如果没有配额, 则请求被拒绝.
func (l SlidingWindowLimiter[B]) Take(ctx context.Context, id string) (*LimiterResult, error) {
	return l.backend.Take(ctx, &LimiterTakeRequest{
		Key:      l.param.formatKey(id),
		Window:   l.param.Window,
		MaxLimit: l.param.MaxLimit,
		UniqueId: UniqueId(),
	})
}

// Check 检查下一个请求是否被允许, 不修改任何数据.
func (l SlidingWindowLimiter[B]) Check(ctx context.Context, id string) (*LimiterResult, error) {
	return l.backend.Check(ctx, &LimiterCheckRequest{
		Key:      l.param.formatKey(id),
		Window:   l.param.Window,
		MaxLimit: l.param.MaxLimit,
	})
}

// Lock 锁定 key, 在滑动窗口内将拒绝所有请求.
func (l SlidingWindowLimiter[B]) Lock(ctx context.Context, id string) (*LimiterResult, error) {
	return l.backend.Lock(ctx, &LimiterLockRequest{
		Key:      l.param.formatKey(id),
		Window:   l.param.Window,
		MaxLimit: l.param.MaxLimit,
	})
}

// Reset 清除 key的所有限制.
func (l SlidingWindowLimiter[B]) Reset(ctx context.Context, id string) error {
	return l.backend.Reset(ctx, l.param.formatKey(id))
}
