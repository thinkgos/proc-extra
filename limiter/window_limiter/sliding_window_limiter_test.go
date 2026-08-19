package window_limiter_test

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	redisv9 "github.com/thinkgos/proc-extra/limiter/window_limiter/redis/v9"
	"github.com/thinkgos/proc-extra/limiter/window_limiter/tests"
)

func Test_SlidingWindowLimiter_Work(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	tests.GenericTest_SlidingWindowLimiter_Work(
		t,
		mr,
		redisv9.NewLimitRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_SlidingWindowLimiter_Lock(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	tests.GenericTest_SlidingWindowLimiter_Lock(
		t,
		mr,
		redisv9.NewLimitRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}
