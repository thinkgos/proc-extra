package limit_verified_test

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	redisV9 "github.com/thinkgos/proc-extra/limiter/limit_verified/redis/v9"
	"github.com/thinkgos/proc-extra/limiter/limit_verified/tests"
)

func Test_LimitVerifiedName(t *testing.T) {
	tests.GenericTest_Name(
		t,
		nil,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})),
	)
}

func Test_LimitVerifiedInvalidScene(t *testing.T) {
	tests.GenericTest_InvalidScene(
		t,
		nil,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})),
	)
}

func Test_LimitVerifiedSuccess(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTest_Work(
		t,
		mr,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_LimitVerifiedSendCode_Failure(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTest_SendCode_Failure(
		t,
		mr,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_LimitVerifiedSendCode_OverQuota(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTest_SendCode_OverQuota(
		t,
		mr,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_LimitVerifiedSendCode_ResendTooFrequently(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTest_SendCode_ResendTooFrequently(
		t,
		mr,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_LimitVerifiedVerifyCode_CodeExpired(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTest_VerifyCode_Expired(
		t,
		mr,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_LimitVerifiedVerifyCode_ReachMaxAttempt(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTest_VerifyCode_ReachMaxAttempt(
		t,
		mr,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}
