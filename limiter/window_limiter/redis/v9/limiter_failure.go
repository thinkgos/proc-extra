package v9

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/thinkgos/proc-extra/limiter/window_limiter"
	redis_script "github.com/thinkgos/proc-extra/limiter/window_limiter/redis"
)

var _ window_limiter.SlidingWindowFailureLimiterBackend = (*LimitFailureRedisStore)(nil)

type LimitFailureRedisStore struct {
	store *redis.Client
}

// NewLimitFailureRedisStore returns a RedisStore with given parameters.
func NewLimitFailureRedisStore(store *redis.Client) *LimitFailureRedisStore {
	return &LimitFailureRedisStore{
		store: store,
	}
}

// Evaluate implements [window_limiter.SlidingWindowFailureLimiterBackend].
func (p *LimitFailureRedisStore) Evaluate(ctx context.Context, v *window_limiter.FailureLimiterEvaluateRequest) (*window_limiter.FailureLimiterResult, error) {
	vals, err := p.store.Eval(ctx,
		redis_script.ScriptSlidingWindowFailureLimiterEvaluate,
		[]string{v.Key},
		[]string{
			strconv.Itoa(v.Window),
			strconv.Itoa(v.MaxFailures),
			v.UniqueId,
			formatBoolString(v.IsFailure),
		},
	).Int64Slice()
	if err != nil {
		return nil, err
	}
	return &window_limiter.FailureLimiterResult{
		Allow:       vals[0] == 0,
		ExpireAt:    vals[1],
		Failures:    int(vals[2]),
		MaxFailures: v.MaxFailures,
	}, nil
}

// Lock implements [window_limiter.SlidingWindowLimiterBackend].
func (p *LimitFailureRedisStore) Lock(ctx context.Context, v *window_limiter.FailureLimiterLockRequest) (*window_limiter.FailureLimiterResult, error) {
	vals, err := p.store.Eval(ctx,
		redis_script.ScriptSlidingWindowFailureLimiterLock,
		[]string{v.Key},
		[]string{
			strconv.Itoa(v.Window),
			strconv.Itoa(v.MaxFailures),
		},
	).Int64Slice()
	if err != nil {
		return nil, err
	}
	return &window_limiter.FailureLimiterResult{
		Allow:       vals[0] == 0,
		ExpireAt:    vals[1],
		Failures:    int(vals[2]),
		MaxFailures: v.MaxFailures,
	}, nil
}

// Reset implements [window_limiter.SlidingWindowLimiterBackend].
func (p *LimitFailureRedisStore) Reset(ctx context.Context, key string) error {
	return p.store.Del(ctx, key, key+":_locked").Err()
}

// Check implements [window_limiter.SlidingWindowLimiterBackend].
func (p *LimitFailureRedisStore) Check(ctx context.Context, v *window_limiter.FailureLimiterCheckRequest) (*window_limiter.FailureLimiterResult, error) {
	vals, err := p.store.Eval(ctx,
		redis_script.ScriptSlidingWindowFailureLimiterCheck,
		[]string{v.Key},
		[]string{
			strconv.Itoa(v.Window),
			strconv.Itoa(v.MaxFailures),
		},
	).Int64Slice()
	if err != nil {
		return nil, err
	}
	return &window_limiter.FailureLimiterResult{
		Allow:       vals[0] == 0,
		ExpireAt:    vals[1],
		Failures:    int(vals[2]),
		MaxFailures: v.MaxFailures,
	}, nil
}

func formatBoolString(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
