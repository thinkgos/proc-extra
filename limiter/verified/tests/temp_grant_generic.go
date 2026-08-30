package tests

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"

	"github.com/thinkgos/proc-extra/limiter/verified"
)

var _ verified.TempGrantGenerator = (*TestTempGrantProvider)(nil)

var testTempGrantSceneParam = &verified.Param{
	KeyExpires:  testKeyExpires,
	MaxAttempts: testMaxAttempts_1,
}

type TestTempGrantProvider struct{}

func (t TestTempGrantProvider) Name() string { return "test-temp-grant-provider" }

func (t TestTempGrantProvider) GenerateUniqueId() string {
	return randString(6)
}

func GenericTest_TempGrant_InMaxAttempts[B verified.StorageBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := verified.NewTempGrant[testSceneType](new(TestTempGrantProvider), backend).
		SetKeyPrefix(testKeyPrefix).
		SetSceneParam(testScene, testTempGrantSceneParam)

	targetId := randString(6)
	wantAnswer, err := l.Issue(
		context.Background(),
		testScene,
		targetId,
		verified.WithMaxAttempts(3),
		verified.WithKeyExpires(time.Minute*5),
	)
	require.NoError(t, err)

	badAnswer := wantAnswer + "xxx"
	b, err := l.Consume(context.Background(), testScene, targetId, badAnswer)
	require.NoError(t, err)
	require.False(t, b)
	b, err = l.Consume(context.Background(), testScene, targetId, badAnswer)
	require.NoError(t, err)
	require.False(t, b)
	b, err = l.Consume(context.Background(), testScene, targetId, wantAnswer)
	require.NoError(t, err)
	require.True(t, b)
	b, err = l.Consume(context.Background(), testScene, targetId, badAnswer)
	require.NoError(t, err)
	require.False(t, b)
}

func GenericTest_TempGrant_OverMaxAttempts[B verified.StorageBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := verified.NewTempGrant[testSceneType](new(TestTempGrantProvider), backend).
		SetKeyPrefix(testKeyPrefix).
		SetSceneParam(testScene, testTempGrantSceneParam)
	targetId := randString(6)
	wantAnswer, err := l.Issue(
		context.Background(),
		testScene,
		targetId,
		verified.WithKeyExpires(time.Minute*3),
		verified.WithMaxAttempts(3),
	)
	require.NoError(t, err)

	badAnswer := wantAnswer + "xxx"
	for range 6 {
		b, err := l.Consume(context.Background(), testScene, targetId, badAnswer)
		require.NoError(t, err)
		require.False(t, b)
	}
	b, err := l.Consume(context.Background(), testScene, targetId, wantAnswer)
	require.NoError(t, err)
	require.False(t, b)
}

func GenericTest_TempGrant_OneShot[B verified.StorageBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := verified.NewTempGrant[testSceneType](new(TestTempGrantProvider), backend).
		SetKeyPrefix(testKeyPrefix).
		SetSceneParam(testScene, testTempGrantSceneParam)

	targetId := randString(6)

	wantAnswer, err := l.Issue(context.Background(), testScene, targetId, verified.WithKeyExpires(time.Minute*5))
	require.NoError(t, err)

	b, err := l.Consume(context.Background(), testScene, targetId, wantAnswer)
	require.NoError(t, err)
	require.True(t, b)

	b, err = l.Consume(context.Background(), testScene, targetId, wantAnswer)
	require.NoError(t, err)
	require.False(t, b)
}

func GenericTest_TempGrant_Validate[B verified.StorageBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := verified.NewTempGrant[testSceneType](new(TestTempGrantProvider), backend).
		SetKeyPrefix(testKeyPrefix).
		SetSceneParam(testScene, testTempGrantSceneParam)

	targetId := randString(6)
	wantAnswer, err := l.Issue(
		context.Background(),
		testScene,
		targetId,
		verified.WithMaxAttempts(3),
		verified.WithKeyExpires(time.Minute*5),
	)
	require.NoError(t, err)

	// valid token returns true
	b, err := l.Validate(context.Background(), testScene, targetId, wantAnswer)
	require.NoError(t, err)
	require.True(t, b)

	// validate is read-only: same token can be validated again
	b, err = l.Validate(context.Background(), testScene, targetId, wantAnswer)
	require.NoError(t, err)
	require.True(t, b)

	// invalid token returns false
	b, err = l.Validate(context.Background(), testScene, targetId, wantAnswer+"xxx")
	require.NoError(t, err)
	require.False(t, b)

	// validate does not consume: valid token still works after invalid attempt
	b, err = l.Validate(context.Background(), testScene, targetId, wantAnswer)
	require.NoError(t, err)
	require.True(t, b)
}

func GenericTest_TempGrant_Validate_Timeout[B verified.StorageBackend](t *testing.T, mr *miniredis.Miniredis, backend B) {
	l := verified.NewTempGrant[testSceneType](new(TestTempGrantProvider), backend).
		SetKeyPrefix(testKeyPrefix).
		SetSceneParam(testScene, testTempGrantSceneParam)

	targetId := randString(6)
	wantAnswer, err := l.Issue(
		context.Background(),
		testScene,
		targetId,
		verified.WithKeyExpires(time.Second*1),
		verified.WithMaxAttempts(3),
	)
	require.NoError(t, err)

	mr.FastForward(time.Second)

	// token expired: validate returns false
	b, err := l.Validate(context.Background(), testScene, targetId, wantAnswer)
	require.NoError(t, err)
	require.False(t, b)
}

func GenericTest_TempGrant_Validate_ThenConsume[B verified.StorageBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := verified.NewTempGrant[testSceneType](new(TestTempGrantProvider), backend).
		SetKeyPrefix(testKeyPrefix).
		SetSceneParam(testScene, testTempGrantSceneParam)

	targetId := randString(6)
	wantAnswer, err := l.Issue(
		context.Background(),
		testScene,
		targetId,
		verified.WithMaxAttempts(3),
		verified.WithKeyExpires(time.Minute*5),
	)
	require.NoError(t, err)

	// validate first (read-only)
	b, err := l.Validate(context.Background(), testScene, targetId, wantAnswer)
	require.NoError(t, err)
	require.True(t, b)

	// consume succeeds after validate
	b, err = l.Consume(context.Background(), testScene, targetId, wantAnswer)
	require.NoError(t, err)
	require.True(t, b)

	// consumed: validate now returns false
	b, err = l.Validate(context.Background(), testScene, targetId, wantAnswer)
	require.NoError(t, err)
	require.False(t, b)
}

func GenericTest_TempGrant_OneShot_Timeout[B verified.StorageBackend](t *testing.T, mr *miniredis.Miniredis, backend B) {
	l := verified.NewTempGrant[testSceneType](new(TestTempGrantProvider), backend).
		SetKeyPrefix(testKeyPrefix).
		SetSceneParam(testScene, testTempGrantSceneParam)
	targetId := randString(6)
	wantAnswer, err := l.Issue(context.Background(), testScene, targetId, verified.WithKeyExpires(time.Second*1))
	require.NoError(t, err)

	mr.FastForward(time.Second)

	b, err := l.Consume(context.Background(), testScene, targetId, wantAnswer)
	require.NoError(t, err)
	require.False(t, b)
}
