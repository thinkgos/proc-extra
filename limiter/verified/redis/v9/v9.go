package v9

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/thinkgos/proc-extra/limiter/verified"
	redisScript "github.com/thinkgos/proc-extra/limiter/verified/redis"
)

var _ verified.StorageBackend = (*RedisStore)(nil)

// RedisStore verified captcha limit
type RedisStore struct {
	client *redis.Client // store redis client
}

// NewRedisStore new redis store instance.
func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
}

// Save the arguments.
func (s *RedisStore) Save(ctx context.Context, p *verified.SaveArgs) error {
	return s.client.Eval(
		ctx,
		redisScript.ScriptSave,
		[]string{p.Key},
		[]string{
			p.Answer,
			strconv.Itoa(p.MaxAttempts),
			strconv.Itoa(int(p.KeyExpires / time.Second)),
		},
	).Err()
}

// Verify the answer.
func (s *RedisStore) Verify(ctx context.Context, p *verified.VerifyArgs) (bool, error) {
	code, err := s.client.Eval(
		ctx,
		redisScript.ScriptVerify,
		[]string{p.Key},
		[]string{p.Answer},
	).Int64()
	if err != nil {
		return false, err
	}
	return code == 0, nil
}

// Inspect the answer.
func (s *RedisStore) Inspect(ctx context.Context, p *verified.VerifyArgs) (bool, error) {
	code, err := s.client.Eval(
		ctx,
		redisScript.ScriptInspect,
		[]string{p.Key},
		[]string{p.Answer},
	).Int64()
	if err != nil {
		return false, err
	}
	return code == 0, nil
}
