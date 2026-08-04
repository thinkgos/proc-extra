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

var _ verified.RefluxProvider = (*TestRefluxProvider)(nil)

var testRefluxParams = map[string]*verified.Param{
	testKind: {
		KeyPrefix:   testKeyPrefix,
		KeyExpires:  testKeyExpires,
		MaxAttempts: testMaxAttempts_1,
	},
}

type TestRefluxProvider struct{}

func (t TestRefluxProvider) Name() string { return "test_provider" }

func (t TestRefluxProvider) GenerateUniqueId() string {
	return randString(6)
}

func GenericTestReflux_Improve_Cover[B verified.StorageBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := verified.NewVerifiedReflux[string, *TestRefluxProvider, B](new(TestRefluxProvider), backend).
		SetParams(testRefluxParams)
	l.Name()
	targetId := randString(6)
	wantAnswer, err := l.Generate(context.Background(), testInvalidKind, targetId)
	require.ErrorIs(t, verified.ErrParamKindNotFound, err)

	_, err = l.Verify(context.Background(), testInvalidKind, targetId, wantAnswer)
	require.ErrorIs(t, verified.ErrParamKindNotFound, err)
}

func GenericTestReflux_One_Time[B verified.StorageBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := verified.NewVerifiedReflux[string, *TestRefluxProvider, B](new(TestRefluxProvider), backend).
		SetParams(testRefluxParams)

	targetId := randString(6)

	wantAnswer, err := l.Generate(context.Background(), testKind, targetId, verified.WithKeyExpires(time.Minute*5))
	assert.NoError(t, err)

	b, err := l.Verify(context.Background(), testKind, targetId, wantAnswer)
	require.NoError(t, err)
	require.True(t, b)

	b, err = l.Verify(context.Background(), testKind, targetId, wantAnswer)
	require.NoError(t, err)
	require.False(t, b)
}

func GenericTestReflux_In_MaxAttempts[B verified.StorageBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := verified.NewVerifiedReflux[string, *TestRefluxProvider, B](new(TestRefluxProvider), backend).
		SetParams(testRefluxParams)

	targetId := randString(6)
	wantAnswer, err := l.Generate(
		context.Background(),
		testKind,
		targetId,
		verified.WithMaxAttempts(3),
		verified.WithKeyExpires(time.Minute*5),
	)
	assert.NoError(t, err)

	badAnswer := wantAnswer + "xxx"
	b, err := l.Verify(context.Background(), testKind, targetId, badAnswer)
	require.NoError(t, err)
	require.False(t, b)
	b, err = l.Verify(context.Background(), testKind, targetId, badAnswer)
	require.NoError(t, err)
	require.False(t, b)
	b, err = l.Verify(context.Background(), testKind, targetId, wantAnswer)
	require.NoError(t, err)
	require.True(t, b)
}

func GenericTestReflux_Over_MaxAttempts[B verified.StorageBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := verified.NewVerifiedReflux[string, *TestRefluxProvider, B](new(TestRefluxProvider), backend).
		SetParams(testRefluxParams)
	targetId := randString(6)
	wantAnswer, err := l.Generate(
		context.Background(),
		testKind,
		targetId,
		verified.WithKeyExpires(time.Minute*3),
		verified.WithMaxAttempts(3),
	)
	assert.NoError(t, err)

	badAnswer := wantAnswer + "xxx"
	for range 6 {
		b, err := l.Verify(context.Background(), testKind, targetId, badAnswer)
		require.NoError(t, err)
		require.False(t, b)
	}
	b, err := l.Verify(context.Background(), testKind, targetId, wantAnswer)
	require.NoError(t, err)
	require.False(t, b)
}

func GenericTestReflux_OneTime_Timeout[B verified.StorageBackend](t *testing.T, mr *miniredis.Miniredis, backend B) {
	l := verified.NewVerifiedReflux[string, *TestRefluxProvider, B](new(TestRefluxProvider), backend).
		SetParams(testRefluxParams)
	targetId := randString(6)
	wantAnswer, err := l.Generate(context.Background(), testKind, targetId, verified.WithKeyExpires(time.Second*1))
	assert.NoError(t, err)

	mr.FastForward(time.Second)

	b, err := l.Verify(context.Background(), testKind, targetId, wantAnswer)
	require.NoError(t, err)
	require.False(t, b)
}
