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
	args := []string{
		strconv.FormatInt(int64(p.Window/time.Second), 10),
		strconv.Itoa(p.Quota),
		strconv.Itoa(p.CodeExpires),
		strconv.Itoa(p.CodeMaxAttempts),
		p.Code,
		p.UniqueId,
		strconv.Itoa(len(p.WindowTiers)),
	}
	for _, tier := range p.WindowTiers {
		args = append(args,
			strconv.FormatInt(int64(tier.Window/time.Second), 10),
			strconv.Itoa(tier.Quota),
		)
	}

	sts, err := v.store.Eval(
		ctx,
		redis_script.ScriptLimitVerifiedEvaluate,
		[]string{
			p.Key,
			p.CodeKey,
		},
		args,
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
			p.Key,
			p.CodeKey,
		},
		[]string{p.UniqueId},
	).Err()
}

// VerifyCode verify code from redis cache.
func (v *RedisStore) Verify(ctx context.Context, p *limit_verified.VerifyRequest) (*limit_verified.VerifyResult, error) {
	sts, err := v.store.Eval(
		ctx,
		redis_script.ScriptLimitVerifiedVerifyCode,
		[]string{
			p.Key,
			p.CodeKey,
		},
		[]string{p.Code},
	).Int64()
	if err != nil {
		return nil, err
	}
	return &limit_verified.VerifyResult{
		Status: limit_verified.VerifyStatus(sts),
	}, nil
}
