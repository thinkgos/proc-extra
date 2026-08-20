package window_limiter

import (
	"context"
)

type WindowLimiter[S SceneValuer] interface {
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

// SlidingWindowLimiter sliding window limiter with scene support.
type SlidingWindowLimiter[S SceneValuer, B SlidingWindowLimiterBackend] struct {
	backend B
	sps     SceneParamRegistry[S, SlidingWindowLimiterParam]
}

// NewSlidingWindowLimiter new sliding window limiter instance.
func NewSlidingWindowLimiter[S SceneValuer, B SlidingWindowLimiterBackend](backend B) *SlidingWindowLimiter[S, B] {
	return &SlidingWindowLimiter[S, B]{
		backend: backend,
		sps: NewSceneParamRegistry[S]("window:limiter:", &SlidingWindowLimiterParam{
			Window:   60,
			MaxLimit: 10,
		}),
	}
}

// SetKeyPrefix sets the key prefix.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (l *SlidingWindowLimiter[S, B]) SetKeyPrefix(keyPrefix string) *SlidingWindowLimiter[S, B] {
	l.sps.SetKeyPrefix(keyPrefix)
	return l
}

// SetParam sets the general param.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (l *SlidingWindowLimiter[S, B]) SetGeneralParam(p *SlidingWindowLimiterParam) *SlidingWindowLimiter[S, B] {
	l.sps.SetGeneralParam(p)
	return l
}

// SetSceneParam sets the param for a specific scene.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (l *SlidingWindowLimiter[S, B]) SetSceneParam(scene S, param *SlidingWindowLimiterParam) *SlidingWindowLimiter[S, B] {
	l.sps.SetSceneParam(scene, param)
	return l
}

// Take 尝试获取一个请求的配额单位.
// 如果有可用配额, 则请求被允许, 并且增加一次配额消费.
// 如果没有配额, 则请求被拒绝.
func (l *SlidingWindowLimiter[S, B]) Take(ctx context.Context, scene S, id string) (*LimiterResult, error) {
	p := l.sps.useScene(scene)
	return l.backend.Take(ctx, &LimiterTakeRequest{
		Key:       l.sps.formatKey(scene.Value(), id),
		LockedKey: l.sps.formatLockedKey(scene.Value(), id),
		Window:    p.Window,
		MaxLimit:  p.MaxLimit,
		UniqueId:  UniqueId(),
	})
}

// Check 检查下一个请求是否被允许, 不修改任何数据.
func (l *SlidingWindowLimiter[S, B]) Check(ctx context.Context, scene S, id string) (*LimiterResult, error) {
	p := l.sps.useScene(scene)
	return l.backend.Check(ctx, &LimiterCheckRequest{
		Key:       l.sps.formatKey(scene.Value(), id),
		LockedKey: l.sps.formatLockedKey(scene.Value(), id),
		Window:    p.Window,
		MaxLimit:  p.MaxLimit,
	})
}

// Lock 锁定 key, 在滑动窗口内将拒绝所有请求.
func (l *SlidingWindowLimiter[S, B]) Lock(ctx context.Context, scene S, id string) (*LimiterResult, error) {
	p := l.sps.useScene(scene)
	return l.backend.Lock(ctx, &LimiterLockRequest{
		Key:       l.sps.formatKey(scene.Value(), id),
		LockedKey: l.sps.formatLockedKey(scene.Value(), id),
		Window:    p.Window,
		MaxLimit:  p.MaxLimit,
	})
}

// Reset 清除 key的所有限制.
func (l *SlidingWindowLimiter[S, B]) Reset(ctx context.Context, scene S, id string) error {
	return l.backend.Reset(ctx, &LimiterResetRequest{
		Key:       l.sps.formatKey(scene.Value(), id),
		LockedKey: l.sps.formatLockedKey(scene.Value(), id),
	})
}
