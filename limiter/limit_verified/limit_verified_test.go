package limit_verified_test

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	redisV9 "github.com/thinkgos/proc-extra/limiter/limit_verified/redis/v9"
	"github.com/thinkgos/proc-extra/limiter/limit_verified/tests"
)

func Test_RedisV9_Name(t *testing.T) {
	tests.GenericTest_Name(
		t,
		nil,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})),
	)
}

func Test_RedisV9_InvalidKind(t *testing.T) {
	tests.GenericTest_InvalidKind(
		t,
		nil,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})),
	)
}

func Test_RedisV9_Success(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTest_Work(
		t,
		mr,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_SendCode_Failure(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestSendCode_Failure(
		t,
		mr,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_SendCode_OverQuota(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestSendCode_OverQuota(
		t,
		mr,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_SendCode_ResendTooFrequently(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestSendCode_ResendTooFrequently(
		t,
		mr,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_VerifyCode_CodeExpired(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestVerifyCode_Expired(
		t,
		mr,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_VerifyCode_ReachMaxAttempt(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestVerifyCode_ReachMaxAttempt(
		t,
		mr,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}
