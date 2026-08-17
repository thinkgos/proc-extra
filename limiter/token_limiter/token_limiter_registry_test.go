package token_limiter_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/thinkgos/proc-extra/limiter/token_limiter"
	v9 "github.com/thinkgos/proc-extra/limiter/token_limiter/redis/v9"
)

func setupRegistry(t *testing.T) (*token_limiter.TokenLimiterRegistry[string, *v9.TokenLimiterStore], *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	t.Cleanup(mr.Close)

	backend := v9.NewTokenLimiterStore(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	registry := token_limiter.NewTokenLimiterRegistry[string, *v9.TokenLimiterStore](backend)
	return registry, mr
}

func Test_NewTokenLimiterRegistry(t *testing.T) {
	registry, _ := setupRegistry(t)
	assert.NotNil(t, registry)
}

func Test_RegistersParams(t *testing.T) {
	registry, _ := setupRegistry(t)

	params := map[string]*token_limiter.Param{
		"scene1": {KeyPrefix: "prefix1:", Rate: 10, Burst: 20},
		"scene2": {KeyPrefix: "prefix2:", Rate: 5, Burst: 10},
	}

	ret := registry.RegistersParams(params)
	assert.Equal(t, registry, ret) // 验证链式调用

	// 验证注册后 Allow 生效
	ok := registry.Allow(context.Background(), "scene1", "id1")
	assert.True(t, ok)
}

func Test_RegistersParams_Merge(t *testing.T) {
	registry, _ := setupRegistry(t)

	registry.RegistersParams(map[string]*token_limiter.Param{
		"scene1": {KeyPrefix: "p1:", Rate: 10, Burst: 20},
	})
	// 再次调用应合并而非覆盖
	registry.RegistersParams(map[string]*token_limiter.Param{
		"scene2": {KeyPrefix: "p2:", Rate: 5, Burst: 10},
	})

	ok1 := registry.Allow(context.Background(), "scene1", "id1")
	ok2 := registry.Allow(context.Background(), "scene2", "id1")
	assert.True(t, ok1)
	assert.True(t, ok2)
}

func Test_RegisterParam(t *testing.T) {
	registry, _ := setupRegistry(t)

	ret := registry.RegisterParam("scene1", &token_limiter.Param{
		KeyPrefix: "prefix1:",
		Rate:      10,
		Burst:     20,
	})
	assert.Equal(t, registry, ret) // 验证链式调用

	ok := registry.Allow(context.Background(), "scene1", "id1")
	assert.True(t, ok)
}

func Test_RegisterParam_Override(t *testing.T) {
	registry, _ := setupRegistry(t)

	registry.RegisterParam("scene1", &token_limiter.Param{
		KeyPrefix: "p1:",
		Rate:      10,
		Burst:     20,
	})
	// 覆盖同一 scene
	registry.RegisterParam("scene1", &token_limiter.Param{
		KeyPrefix: "p2:",
		Rate:      5,
		Burst:     10,
	})

	ok := registry.Allow(context.Background(), "scene1", "id1")
	assert.True(t, ok)
}

func Test_Allow_SceneNotFound(t *testing.T) {
	registry, _ := setupRegistry(t)

	ok := registry.Allow(context.Background(), "nonexistent", "id1")
	assert.False(t, ok)
}

func Test_Allow_ValidScene(t *testing.T) {
	registry, mr := setupRegistry(t)

	registry.RegisterParam("scene1", &token_limiter.Param{
		KeyPrefix: "tokenlimit:",
		Rate:      5,
		Burst:     10,
	})

	ok := registry.Allow(context.Background(), "scene1", "user1")
	assert.True(t, ok)

	// 验证 key 格式正确
	keys := mr.Keys()
	assert.True(t, len(keys) > 0)
}

func Test_Allow_BurstExhausted(t *testing.T) {
	registry, _ := setupRegistry(t)

	registry.RegisterParam("scene1", &token_limiter.Param{
		KeyPrefix: "tokenlimit:",
		Rate:      1,
		Burst:     2,
	})

	// burst=2，初始可用 2 个令牌
	assert.True(t, registry.Allow(context.Background(), "scene1", "user1"))
	assert.True(t, registry.Allow(context.Background(), "scene1", "user1"))
	// burst 已耗尽，后续应拒绝
	assert.False(t, registry.Allow(context.Background(), "scene1", "user1"))
}

func Test_AllowN_SceneNotFound(t *testing.T) {
	registry, _ := setupRegistry(t)

	ok := registry.AllowN(context.Background(), "nonexistent", "id1", 1)
	assert.False(t, ok)
}

func Test_AllowN_ValidScene(t *testing.T) {
	registry, _ := setupRegistry(t)

	registry.RegisterParam("scene1", &token_limiter.Param{
		KeyPrefix: "tokenlimit:",
		Rate:      5,
		Burst:     10,
	})

	ok := registry.AllowN(context.Background(), "scene1", "user1", 1)
	assert.True(t, ok)
}

func Test_AllowN_BurstExhausted(t *testing.T) {
	registry, _ := setupRegistry(t)

	registry.RegisterParam("scene1", &token_limiter.Param{
		KeyPrefix: "tokenlimit:",
		Rate:      1,
		Burst:     5,
	})

	// 一次请求 5 个令牌，耗尽 burst
	ok := registry.AllowN(context.Background(), "scene1", "user1", 5)
	assert.True(t, ok)
	// 再请求 1 个应拒绝
	ok = registry.AllowN(context.Background(), "scene1", "user1", 1)
	assert.False(t, ok)
}

func Test_AllowN_ExceedBurst(t *testing.T) {
	registry, _ := setupRegistry(t)

	registry.RegisterParam("scene1", &token_limiter.Param{
		KeyPrefix: "tokenlimit:",
		Rate:      5,
		Burst:     3,
	})

	// 请求数超过 burst
	ok := registry.AllowN(context.Background(), "scene1", "user1", 10)
	assert.False(t, ok)
}

func Test_Allow_MultipleScenes(t *testing.T) {
	registry, _ := setupRegistry(t)

	registry.RegisterParam("scene1", &token_limiter.Param{
		KeyPrefix: "s1:",
		Rate:      5,
		Burst:     10,
	}).RegisterParam("scene2", &token_limiter.Param{
		KeyPrefix: "s2:",
		Rate:      3,
		Burst:     5,
	})

	// 不同 scene 互不影响
	ok1 := registry.Allow(context.Background(), "scene1", "user1")
	ok2 := registry.Allow(context.Background(), "scene2", "user1")
	assert.True(t, ok1)
	assert.True(t, ok2)
}

func Test_Allow_DifferentUsers(t *testing.T) {
	registry, _ := setupRegistry(t)

	registry.RegisterParam("scene1", &token_limiter.Param{
		KeyPrefix: "tokenlimit:",
		Rate:      1,
		Burst:     1,
	})

	// 不同用户各自独立计数
	ok1 := registry.Allow(context.Background(), "scene1", "user1")
	ok2 := registry.Allow(context.Background(), "scene1", "user2")
	assert.True(t, ok1)
	assert.True(t, ok2)
}

func Test_Chaining(t *testing.T) {
	registry, _ := setupRegistry(t)

	// 验证链式调用
	ret := registry.
		RegisterParam("s1", &token_limiter.Param{KeyPrefix: "p1:", Rate: 1, Burst: 1}).
		RegisterParam("s2", &token_limiter.Param{KeyPrefix: "p2:", Rate: 1, Burst: 1}).
		RegistersParams(map[string]*token_limiter.Param{
			"s3": {KeyPrefix: "p3:", Rate: 1, Burst: 1},
		})

	assert.Equal(t, registry, ret)
	assert.True(t, registry.Allow(context.Background(), "s1", "id"))
	assert.True(t, registry.Allow(context.Background(), "s2", "id"))
	assert.True(t, registry.Allow(context.Background(), "s3", "id"))
}

func Test_TryAllow_SceneNotFound(t *testing.T) {
	registry, _ := setupRegistry(t)

	ok, err := registry.TryAllow(context.Background(), "nonexistent", "id1")
	assert.False(t, ok)
	assert.ErrorIs(t, err, token_limiter.ErrSceneParamNotFound)
}

func Test_TryAllow_ValidScene(t *testing.T) {
	registry, _ := setupRegistry(t)

	registry.RegisterParam("scene1", &token_limiter.Param{
		KeyPrefix: "tokenlimit:",
		Rate:      5,
		Burst:     10,
	})

	ok, err := registry.TryAllow(context.Background(), "scene1", "user1")
	assert.True(t, ok)
	assert.NoError(t, err)
}

func Test_TryAllow_BurstExhausted(t *testing.T) {
	registry, _ := setupRegistry(t)

	registry.RegisterParam("scene1", &token_limiter.Param{
		KeyPrefix: "tokenlimit:",
		Rate:      1,
		Burst:     2,
	})

	assert.True(t, must(registry.TryAllow(context.Background(), "scene1", "user1")))
	assert.True(t, must(registry.TryAllow(context.Background(), "scene1", "user1")))
	// burst 已耗尽
	ok, err := registry.TryAllow(context.Background(), "scene1", "user1")
	assert.False(t, ok)
	assert.NoError(t, err)
}

func Test_TryAllowN_SceneNotFound(t *testing.T) {
	registry, _ := setupRegistry(t)

	ok, err := registry.TryAllowN(context.Background(), "nonexistent", "id1", 1)
	assert.False(t, ok)
	assert.ErrorIs(t, err, token_limiter.ErrSceneParamNotFound)
}

func Test_TryAllowN_ValidScene(t *testing.T) {
	registry, _ := setupRegistry(t)

	registry.RegisterParam("scene1", &token_limiter.Param{
		KeyPrefix: "tokenlimit:",
		Rate:      5,
		Burst:     10,
	})

	ok, err := registry.TryAllowN(context.Background(), "scene1", "user1", 1)
	assert.True(t, ok)
	assert.NoError(t, err)
}

func Test_TryAllowN_ExceedBurst(t *testing.T) {
	registry, _ := setupRegistry(t)

	registry.RegisterParam("scene1", &token_limiter.Param{
		KeyPrefix: "tokenlimit:",
		Rate:      5,
		Burst:     3,
	})

	// 请求数超过 burst
	ok, err := registry.TryAllowN(context.Background(), "scene1", "user1", 10)
	assert.False(t, ok)
	assert.NoError(t, err)
}

func Test_AllowNAt_BurstExhausted(t *testing.T) {
	registry, _ := setupRegistry(t)

	registry.RegisterParam("scene1", &token_limiter.Param{
		KeyPrefix: "tokenlimit:",
		Rate:      1,
		Burst:     5,
	})

	now := time.Now()
	// 一次请求 5 个令牌，耗尽 burst
	ok := registry.AllowNAt(context.Background(), "scene1", "user1", 5, now)
	assert.True(t, ok)
	// 再请求 1 个应拒绝
	ok = registry.AllowNAt(context.Background(), "scene1", "user1", 1, now)
	assert.False(t, ok)
}

func Test_TryAllowNAt_SceneNotFound(t *testing.T) {
	registry, _ := setupRegistry(t)

	ok, err := registry.TryAllowNAt(context.Background(), "nonexistent", "id1", 1, time.Now())
	assert.False(t, ok)
	assert.ErrorIs(t, err, token_limiter.ErrSceneParamNotFound)
}

func Test_TryAllowNAt_ValidScene(t *testing.T) {
	registry, _ := setupRegistry(t)

	registry.RegisterParam("scene1", &token_limiter.Param{
		KeyPrefix: "tokenlimit:",
		Rate:      5,
		Burst:     10,
	})

	ok, err := registry.TryAllowNAt(context.Background(), "scene1", "user1", 1, time.Now())
	assert.True(t, ok)
	assert.NoError(t, err)
}

func Test_TryAllowNAt_ExceedBurst(t *testing.T) {
	registry, _ := setupRegistry(t)

	registry.RegisterParam("scene1", &token_limiter.Param{
		KeyPrefix: "tokenlimit:",
		Rate:      5,
		Burst:     3,
	})

	// 请求数超过 burst
	ok, err := registry.TryAllowNAt(context.Background(), "scene1", "user1", 10, time.Now())
	assert.False(t, ok)
	assert.NoError(t, err)
}

func must(ok bool, err error) bool {
	if err != nil {
		panic(err)
	}
	return ok
}
