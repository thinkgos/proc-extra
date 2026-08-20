package token_limiter_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thinkgos/proc-extra/limiter/token_limiter"
	v9 "github.com/thinkgos/proc-extra/limiter/token_limiter/redis/v9"
)

type testScene string

func (s testScene) Value() string { return string(s) }

const (
	sceneNormal testScene = "normal"
	sceneBurst  testScene = "burst"
	sceneStrict testScene = "strict"
)

func setupTokenLimiter(t *testing.T) (*token_limiter.TokenLimiter[testScene, *v9.TokenLimiterStore], *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	backend := v9.NewTokenLimiterStore(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	tl := token_limiter.NewTokenLimiter[testScene, *v9.TokenLimiterStore](backend)
	return tl, mr
}

func Test_NewTokenLimiter(t *testing.T) {
	tl, _ := setupTokenLimiter(t)
	assert.NotNil(t, tl)
}

func Test_SetKeyPrefix(t *testing.T) {
	tl, mr := setupTokenLimiter(t)

	tl.SetKeyPrefix("custom:prefix:").
		SetGeneralParam(&token_limiter.Param{Rate: 5, Burst: 10})

	ok := tl.Allow(context.Background(), sceneNormal, "user1")
	assert.True(t, ok)

	keys := mr.Keys()
	found := false
	for _, k := range keys {
		if k == "custom:prefix:normal:user1" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected key with custom prefix")
}

func Test_SetGeneralParam(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	// Set general param with very small burst
	tl.SetGeneralParam(&token_limiter.Param{Rate: 1, Burst: 1})

	ok := tl.Allow(context.Background(), sceneNormal, "user1")
	assert.True(t, ok)

	// burst=1 exhausted
	ok = tl.Allow(context.Background(), sceneNormal, "user1")
	assert.False(t, ok)
}

func Test_SetSceneParam(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	tl.SetSceneParam(sceneNormal, &token_limiter.Param{Rate: 5, Burst: 10})

	ok := tl.Allow(context.Background(), sceneNormal, "user1")
	assert.True(t, ok)
}

func Test_SetSceneParam_Override(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	tl.SetSceneParam(sceneNormal, &token_limiter.Param{Rate: 10, Burst: 20})
	// Override same scene
	tl.SetSceneParam(sceneNormal, &token_limiter.Param{Rate: 1, Burst: 1})

	// first consume ok
	ok := tl.Allow(context.Background(), sceneNormal, "user1")
	assert.True(t, ok)
	// burst=1 exhausted
	ok = tl.Allow(context.Background(), sceneNormal, "user1")
	assert.False(t, ok)
}

func Test_SetSceneParam_FallbackToGeneral(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	tl.SetGeneralParam(&token_limiter.Param{Rate: 1, Burst: 1})

	// sceneNormal not registered as scene, falls back to general param
	ok := tl.Allow(context.Background(), sceneNormal, "user1")
	assert.True(t, ok)
	// burst=1 exhausted
	ok = tl.Allow(context.Background(), sceneNormal, "user1")
	assert.False(t, ok)
}

// --- Allow ---

func Test_Allow_Basic(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	tl.SetSceneParam(sceneNormal, &token_limiter.Param{Rate: 5, Burst: 10})

	ok := tl.Allow(context.Background(), sceneNormal, "user1")
	assert.True(t, ok)
}

func Test_Allow_BurstExhausted(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	tl.SetSceneParam(sceneBurst, &token_limiter.Param{Rate: 1, Burst: 2})

	// burst=2, initially 2 tokens available
	assert.True(t, tl.Allow(context.Background(), sceneBurst, "user1"))
	assert.True(t, tl.Allow(context.Background(), sceneBurst, "user1"))
	// burst exhausted, subsequent should deny
	assert.False(t, tl.Allow(context.Background(), sceneBurst, "user1"))
}

func Test_Allow_DifferentUsers(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	tl.SetSceneParam(sceneNormal, &token_limiter.Param{Rate: 1, Burst: 1})

	// Different users have independent counts
	ok1 := tl.Allow(context.Background(), sceneNormal, "user1")
	ok2 := tl.Allow(context.Background(), sceneNormal, "user2")
	assert.True(t, ok1)
	assert.True(t, ok2)
}

func Test_Allow_DifferentScenes(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	tl.SetSceneParam(sceneNormal, &token_limiter.Param{Rate: 5, Burst: 10}).
		SetSceneParam(sceneStrict, &token_limiter.Param{Rate: 1, Burst: 1})

	// Different scenes are independent
	ok1 := tl.Allow(context.Background(), sceneNormal, "user1")
	ok2 := tl.Allow(context.Background(), sceneStrict, "user1")
	assert.True(t, ok1)
	assert.True(t, ok2)
}

// --- AllowN ---

func Test_AllowN_Basic(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	tl.SetSceneParam(sceneNormal, &token_limiter.Param{Rate: 5, Burst: 10})

	ok := tl.AllowN(context.Background(), sceneNormal, "user1", 3)
	assert.True(t, ok)
}

func Test_AllowN_BurstExhausted(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	tl.SetSceneParam(sceneBurst, &token_limiter.Param{Rate: 1, Burst: 5})

	// consume 5 tokens at once, exhaust burst
	ok := tl.AllowN(context.Background(), sceneBurst, "user1", 5)
	assert.True(t, ok)
	// request 1 more should deny
	ok = tl.AllowN(context.Background(), sceneBurst, "user1", 1)
	assert.False(t, ok)
}

func Test_AllowN_ExceedBurst(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	tl.SetSceneParam(sceneNormal, &token_limiter.Param{Rate: 5, Burst: 3})

	// request exceeds burst
	ok := tl.AllowN(context.Background(), sceneNormal, "user1", 10)
	assert.False(t, ok)
}

// --- TryAllow ---

func Test_TryAllow_Basic(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	tl.SetSceneParam(sceneNormal, &token_limiter.Param{Rate: 5, Burst: 10})

	ok, err := tl.TryAllow(context.Background(), sceneNormal, "user1")
	assert.True(t, ok)
	assert.NoError(t, err)
}

func Test_TryAllow_BurstExhausted(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	tl.SetSceneParam(sceneBurst, &token_limiter.Param{Rate: 1, Burst: 2})

	ok, err := tl.TryAllow(context.Background(), sceneBurst, "user1")
	assert.True(t, ok)
	assert.NoError(t, err)

	ok, err = tl.TryAllow(context.Background(), sceneBurst, "user1")
	assert.True(t, ok)
	assert.NoError(t, err)

	// burst exhausted
	ok, err = tl.TryAllow(context.Background(), sceneBurst, "user1")
	assert.False(t, ok)
	assert.NoError(t, err)
}

func Test_TryAllow_FallbackToGeneral(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	tl.SetGeneralParam(&token_limiter.Param{Rate: 1, Burst: 1})

	// unregistered scene falls back to general
	ok, err := tl.TryAllow(context.Background(), sceneNormal, "user1")
	assert.True(t, ok)
	assert.NoError(t, err)

	ok, err = tl.TryAllow(context.Background(), sceneNormal, "user1")
	assert.False(t, ok)
	assert.NoError(t, err)
}

// --- TryAllowN ---

func Test_TryAllowN_Basic(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	tl.SetSceneParam(sceneNormal, &token_limiter.Param{Rate: 5, Burst: 10})

	ok, err := tl.TryAllowN(context.Background(), sceneNormal, "user1", 3)
	assert.True(t, ok)
	assert.NoError(t, err)
}

func Test_TryAllowN_ExceedBurst(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	tl.SetSceneParam(sceneNormal, &token_limiter.Param{Rate: 5, Burst: 3})

	ok, err := tl.TryAllowN(context.Background(), sceneNormal, "user1", 10)
	assert.False(t, ok)
	assert.NoError(t, err)
}

// --- AllowAt / AllowNAt (client-supplied time) ---

func Test_AllowAt_Basic(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	tl.SetSceneParam(sceneNormal, &token_limiter.Param{Rate: 5, Burst: 10})

	ok := tl.AllowAt(context.Background(), sceneNormal, "user1", time.Now())
	assert.True(t, ok)
}

func Test_AllowNAt_BurstExhausted(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	tl.SetSceneParam(sceneBurst, &token_limiter.Param{Rate: 1, Burst: 5})

	now := time.Now()
	// consume 5 tokens, exhaust burst
	ok := tl.AllowNAt(context.Background(), sceneBurst, "user1", 5, now)
	assert.True(t, ok)
	// request 1 more should deny
	ok = tl.AllowNAt(context.Background(), sceneBurst, "user1", 1, now)
	assert.False(t, ok)
}

func Test_AllowNAt_TimeRefill(t *testing.T) {
	tl, mr := setupTokenLimiter(t)

	tl.SetSceneParam(sceneStrict, &token_limiter.Param{Rate: 10, Burst: 5})

	now := time.Now()
	// exhaust burst
	ok := tl.AllowNAt(context.Background(), sceneStrict, "user1", 5, now)
	assert.True(t, ok)
	ok = tl.AllowNAt(context.Background(), sceneStrict, "user1", 1, now)
	assert.False(t, ok)

	// advance time enough to refill tokens (1 second at rate=10)
	mr.FastForward(time.Second)
	later := now.Add(time.Second)
	ok = tl.AllowNAt(context.Background(), sceneStrict, "user1", 1, later)
	assert.True(t, ok)
}

// --- TryAllowAt / TryAllowNAt ---

func Test_TryAllowAt_Basic(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	tl.SetSceneParam(sceneNormal, &token_limiter.Param{Rate: 5, Burst: 10})

	ok, err := tl.TryAllowAt(context.Background(), sceneNormal, "user1", time.Now())
	assert.True(t, ok)
	assert.NoError(t, err)
}

func Test_TryAllowNAt_Basic(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	tl.SetSceneParam(sceneNormal, &token_limiter.Param{Rate: 5, Burst: 10})

	ok, err := tl.TryAllowNAt(context.Background(), sceneNormal, "user1", 3, time.Now())
	assert.True(t, ok)
	assert.NoError(t, err)
}

func Test_TryAllowNAt_ExceedBurst(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	tl.SetSceneParam(sceneNormal, &token_limiter.Param{Rate: 5, Burst: 3})

	ok, err := tl.TryAllowNAt(context.Background(), sceneNormal, "user1", 10, time.Now())
	assert.False(t, ok)
	assert.NoError(t, err)
}

func Test_TryAllowNAt_FallbackToGeneral(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	tl.SetGeneralParam(&token_limiter.Param{Rate: 1, Burst: 1})

	// unregistered scene falls back to general
	ok, err := tl.TryAllowNAt(context.Background(), sceneNormal, "user1", 1, time.Now())
	assert.True(t, ok)
	assert.NoError(t, err)

	ok, err = tl.TryAllowNAt(context.Background(), sceneNormal, "user1", 1, time.Now())
	assert.False(t, ok)
	assert.NoError(t, err)
}

// --- Chaining ---

func Test_Chaining(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	ret := tl.
		SetKeyPrefix("chain:test:").
		SetGeneralParam(&token_limiter.Param{Rate: 100, Burst: 200}).
		SetSceneParam(sceneNormal, &token_limiter.Param{Rate: 5, Burst: 10}).
		SetSceneParam(sceneStrict, &token_limiter.Param{Rate: 1, Burst: 1})

	assert.Equal(t, tl, ret)
	assert.True(t, tl.Allow(context.Background(), sceneNormal, "id"))
	assert.True(t, tl.Allow(context.Background(), sceneStrict, "id"))
}

// --- Rate interface compliance ---

func Test_Rate_Interface(t *testing.T) {
	tl, _ := setupTokenLimiter(t)

	// Verify TokenLimiter implements Rate
	var _ token_limiter.Rate[testScene] = tl
}
