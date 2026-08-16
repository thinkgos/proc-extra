package v9

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"

	"github.com/thinkgos/proc-extra/limiter/token_limiter"
	redis_script "github.com/thinkgos/proc-extra/limiter/token_limiter/redis"
)

var _ token_limiter.TokenLimiterBackend = (*TokenLimiterStore)(nil)

// TokenLimiterStore controls how frequently events are allowed to happen with in one second.
type TokenLimiterStore struct {
	client *redis.Client
}

// NewTokenLimiterStore returns a new TokenLimit that allows events up to rate and permits
// bursts of at most burst tokens.
func NewTokenLimiterStore(client *redis.Client) *TokenLimiterStore {
	return &TokenLimiterStore{
		client: client,
	}
}

// AllowN reports whether n events may happen at time now.
// Use this method if you intend to drop / skip events that exceed the rate.
func (t *TokenLimiterStore) AllowN(ctx context.Context, r *token_limiter.AllowNRequest) (bool, error) {
	resp, err := t.client.Eval(ctx,
		redis_script.ScriptTokenLimiter,
		[]string{
			r.Key,
		},
		[]string{
			strconv.Itoa(r.Rate),
			strconv.Itoa(r.Burst),
			strconv.FormatInt(r.Now.Unix(), 10),
			strconv.Itoa(r.N),
		}).Result()
	// redis allowed == false
	// Lua boolean false -> r Nil bulk reply
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	code, ok := resp.(int64)
	if !ok {
		return false, nil
	}
	// redis allowed == true
	// Lua boolean true -> r integer reply with value of 1
	return code == 1, nil
}
