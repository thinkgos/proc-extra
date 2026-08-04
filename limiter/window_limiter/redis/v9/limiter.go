package v9

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/thinkgos/proc-extra/limiter/window_limiter"
	redis_script "github.com/thinkgos/proc-extra/limiter/window_limiter/redis"
)

var _ window_limiter.SlidingWindowLimiterBackend = (*LimitRedisStore)(nil)

type LimitRedisStore struct {
	store *redis.Client
}

// NewLimitRedisStore returns a RedisStore with given parameters.
func NewLimitRedisStore(store *redis.Client) *LimitRedisStore {
	return &LimitRedisStore{
		store: store,
	}
}

// Take implements [window_limiter.SlidingWindowLimiterBackend].
func (p *LimitRedisStore) Take(ctx context.Context, v *window_limiter.LimiterTakeRequest) (*window_limiter.LimiterResult, error) {
	vals, err := p.store.Eval(ctx,
		redis_script.ScriptSlidingWindowLimiterTake,
		[]string{v.Key},
		[]string{
			strconv.Itoa(v.Window),
			strconv.Itoa(v.MaxLimit),
			v.UniqueId,
		},
	).Int64Slice()
	if err != nil {
		return nil, err
	}
	return &window_limiter.LimiterResult{
		Allow:    vals[0] == 0,
		ExpireAt: vals[1],
		Count:    int(vals[2]),
		MaxLimit: v.MaxLimit,
	}, nil
}

// Lock implements [window_limiter.SlidingWindowLimiterBackend].
func (p *LimitRedisStore) Lock(ctx context.Context, v *window_limiter.LimiterLockRequest) (*window_limiter.LimiterResult, error) {
	vals, err := p.store.Eval(ctx,
		redis_script.ScriptSlidingWindowLimiterLock,
		[]string{v.Key},
		[]string{
			strconv.Itoa(v.Window),
			strconv.Itoa(v.MaxLimit),
		},
	).Int64Slice()
	if err != nil {
		return nil, err
	}
	return &window_limiter.LimiterResult{
		Allow:    vals[0] == 0,
		ExpireAt: vals[1],
		Count:    int(vals[2]),
		MaxLimit: v.MaxLimit,
	}, nil
}

// Reset implements [window_limiter.SlidingWindowLimiterBackend].
func (p *LimitRedisStore) Reset(ctx context.Context, key string) error {
	return p.store.Del(ctx, key, key+":_locked").Err()
}

// Check implements [window_limiter.SlidingWindowLimiterBackend].
func (p *LimitRedisStore) Check(ctx context.Context, v *window_limiter.LimiterCheckRequest) (*window_limiter.LimiterResult, error) {
	vals, err := p.store.Eval(ctx,
		redis_script.ScriptSlidingWindowLimiterCheck,
		[]string{v.Key},
		[]string{
			strconv.Itoa(v.Window),
			strconv.Itoa(v.MaxLimit),
		},
	).Int64Slice()
	if err != nil {
		return nil, err
	}
	return &window_limiter.LimiterResult{
		Allow:    int(vals[0]) == 0,
		ExpireAt: vals[1],
		Count:    int(vals[2]),
		MaxLimit: v.MaxLimit,
	}, nil
}
