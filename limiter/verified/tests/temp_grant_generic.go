package tests

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/thinkgos/proc-extra/limiter/verified"
)

var _ verified.TempGrantGenerator = (*TestTempGrantProvider)(nil)

var testTempGrantParams = map[string]*verified.Param{
	testScene: {
		KeyPrefix:   testKeyPrefix,
		KeyExpires:  testKeyExpires,
		MaxAttempts: testMaxAttempts_1,
	},
}

type TestTempGrantProvider struct{}

func (t TestTempGrantProvider) Name() string { return "test-temp-grant-provider" }

func (t TestTempGrantProvider) GenerateUniqueId() string {
	return randString(6)
}

func GenericTest_TempGrant_ImproveCoverage[B verified.StorageBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := verified.NewTempGrantRegistry[string](new(TestTempGrantProvider), backend).
		SetParams(testTempGrantParams)
	l.Name()
	targetId := randString(6)
	wantAnswer, err := l.Issue(context.Background(), testInvalidScene, targetId)
	require.ErrorIs(t, verified.ErrSceneParamNotFound, err)

	_, err = l.Consume(context.Background(), testInvalidScene, targetId, wantAnswer)
	require.ErrorIs(t, verified.ErrSceneParamNotFound, err)
}

func GenericTest_TempGrant_InMaxAttempts[B verified.StorageBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := verified.NewTempGrantRegistry[string](new(TestTempGrantProvider), backend).
		SetParams(testTempGrantParams)

	targetId := randString(6)
	wantAnswer, err := l.Issue(
		context.Background(),
		testScene,
		targetId,
		verified.WithMaxAttempts(3),
		verified.WithKeyExpires(time.Minute*5),
	)
	assert.NoError(t, err)

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
	l := verified.NewTempGrantRegistry[string](new(TestTempGrantProvider), backend).
		SetParams(testTempGrantParams)
	targetId := randString(6)
	wantAnswer, err := l.Issue(
		context.Background(),
		testScene,
		targetId,
		verified.WithKeyExpires(time.Minute*3),
		verified.WithMaxAttempts(3),
	)
	assert.NoError(t, err)

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
	l := verified.NewTempGrantRegistry[string](new(TestTempGrantProvider), backend).
		SetParams(testTempGrantParams)

	targetId := randString(6)

	wantAnswer, err := l.Issue(context.Background(), testScene, targetId, verified.WithKeyExpires(time.Minute*5))
	assert.NoError(t, err)

	b, err := l.Consume(context.Background(), testScene, targetId, wantAnswer)
	require.NoError(t, err)
	require.True(t, b)

	b, err = l.Consume(context.Background(), testScene, targetId, wantAnswer)
	require.NoError(t, err)
	require.False(t, b)
}

func GenericTest_TempGrant_OneShot_Timeout[B verified.StorageBackend](t *testing.T, mr *miniredis.Miniredis, backend B) {
	l := verified.NewTempGrantRegistry[string](new(TestTempGrantProvider), backend).
		SetParams(testTempGrantParams)
	targetId := randString(6)
	wantAnswer, err := l.Issue(context.Background(), testScene, targetId, verified.WithKeyExpires(time.Second*1))
	assert.NoError(t, err)

	mr.FastForward(time.Second)

	b, err := l.Consume(context.Background(), testScene, targetId, wantAnswer)
	require.NoError(t, err)
	require.False(t, b)
}
