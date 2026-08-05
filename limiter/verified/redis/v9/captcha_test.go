package v9

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thinkgos/proc-extra/limiter/verified/tests"
)

func Test_Captcha_ImproveCoverage(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()
	tests.GenericTest_Captcha_ImproveCoverage(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_Captcha_UnsupportedChallengeProvider(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	addr := mr.Addr()
	mr.Close()
	tests.GenericTest_Captcha_UnsupportedChallengeProvider(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: addr})),
	)
}

func Test_Captcha_InMaxAttempts(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	tests.GenericTest_Captcha_InMaxAttempts(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_Captcha_OverMaxAttempts(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	tests.GenericTest_Captcha_OverMaxAttempts(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}
func Test_Captcha_OneShot(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	tests.GenericTest_Captcha_OneShot(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}
func Test_Captcha_OneShot_Timeout(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	tests.GenericTest_Captcha_OneShot_Timeout(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}
