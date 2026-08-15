package verified_test

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	redisV9 "github.com/thinkgos/proc-extra/limiter/verified/redis/v9"
	"github.com/thinkgos/proc-extra/limiter/verified/tests"
)

func Test_TempGrant_ImproveCoverage(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()
	tests.GenericTest_TempGrant_ImproveCoverage(
		t,
		mr,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_TempGrant_InMaxAttempts(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	tests.GenericTest_TempGrant_InMaxAttempts(
		t,
		mr,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_TempGrant_OverMaxAttempts(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	tests.GenericTest_TempGrant_OverMaxAttempts(
		t,
		mr,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}
func Test_TempGrant_OneShot(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	tests.GenericTest_TempGrant_OneShot(
		t,
		mr,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}
func Test_TempGrant_OneShot_Timeout(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	tests.GenericTest_TempGrant_OneShot_Timeout(
		t,
		mr,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}
