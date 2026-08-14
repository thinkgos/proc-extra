package v9

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/thinkgos/proc-extra/limiter/limit_verified"
	redis_script "github.com/thinkgos/proc-extra/limiter/limit_verified/redis"
)

// RedisStore verified captcha limit
type RedisStore struct {
	store *redis.Client // store client
}

// NewRedisStore
func NewRedisStore(store *redis.Client) *RedisStore {
	return &RedisStore{store}
}

func (v *RedisStore) Evaluate(ctx context.Context, p *limit_verified.EvaluateRequest) (*limit_verified.EvaluateResult, error) {
	sts, err := v.store.Eval(
		ctx,
		redis_script.ScriptLimitVerifiedEvaluate,
		[]string{
			p.KeyPrefix,
			p.Target,
			p.Scene,
		},
		[]string{
			strconv.FormatInt(int64(p.Window/time.Second), 10),
			strconv.Itoa(p.Quota),
			strconv.Itoa(p.ResendInterval),
			strconv.Itoa(p.CodeExpires),
			strconv.Itoa(p.CodeMaxAttempts),
			p.Code,
			p.UniqueId,
		},
	).Int64()
	if err != nil {
		return nil, err
	}
	return &limit_verified.EvaluateResult{
		Status: limit_verified.EvaluateStatus(sts),
	}, nil
}

func (v *RedisStore) Rollback(ctx context.Context, p *limit_verified.RollbackRequest) error {
	return v.store.Eval(
		ctx,
		redis_script.ScriptLimitVerifiedRollback,
		[]string{
			p.KeyPrefix,
			p.Target,
			p.Scene,
		},
		[]string{
			p.UniqueId,
		},
	).Err()
}

// VerifyCode verify code from redis cache.
func (v *RedisStore) Verify(ctx context.Context, p *limit_verified.VerifyRequest) (*limit_verified.VerifyResult, error) {
	sts, err := v.store.Eval(
		ctx,
		redis_script.LimitVerifiedVerifyCodeScript,
		[]string{
			p.KeyPrefix,
			p.Target,
			p.Scene,
		},
		[]string{
			p.Code,
		},
	).Int64()
	if err != nil {
		return nil, err
	}
	return &limit_verified.VerifyResult{
		Status: limit_verified.VerifyStatus(sts),
	}, nil
}
