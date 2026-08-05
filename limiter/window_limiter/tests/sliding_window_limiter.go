package tests

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"
	"github.com/thinkgos/proc-extra/limiter/window_limiter"
)

const testSlidingWindowLimiterScene_Invalid = "invalid_kind"
const testSlidingWindowLimiterScene = "test_kind"

const (
	testSlidingWindowLimiterKeyPrefix = "period:limiter:test-scene:"
	testSlidingWindowLimiterWindow    = 10
	testSlidingWindowLimiterMaxLimit  = 3
)

const testSlidingWindowLimiterId1 = "id1"

var testSlidingWindowLimiterParams = map[string]*window_limiter.SlidingWindowLimiterParam{
	testSlidingWindowLimiterScene: {
		KeyPrefix: testSlidingWindowLimiterKeyPrefix,
		Window:    testSlidingWindowLimiterWindow,
		MaxLimit:  testSlidingWindowLimiterMaxLimit,
	},
}

func GenericTest_SlidingWindowLimiter_InvalidScene[B window_limiter.SlidingWindowLimiterBackend](t *testing.T, mr *miniredis.Miniredis, backend B) {
	l := window_limiter.NewSlidingWindowLimiter[string](backend).
		SetParams(testSlidingWindowLimiterParams)

	_, err := l.Take(context.Background(), testSlidingWindowLimiterScene_Invalid, testSlidingWindowLimiterId1)
	require.ErrorIs(t, window_limiter.ErrSceneParamNotFound, err)

	_, err = l.Check(context.Background(), testSlidingWindowLimiterScene_Invalid, testSlidingWindowLimiterId1)
	require.ErrorIs(t, window_limiter.ErrSceneParamNotFound, err)

	_, err = l.Lock(context.Background(), testSlidingWindowLimiterScene_Invalid, testSlidingWindowLimiterId1)
	require.ErrorIs(t, window_limiter.ErrSceneParamNotFound, err)

	err = l.Reset(context.Background(), testSlidingWindowLimiterScene_Invalid, testSlidingWindowLimiterId1)
	require.ErrorIs(t, window_limiter.ErrSceneParamNotFound, err)
}

func GenericTest_SlidingWindowLimiter_Work[B window_limiter.SlidingWindowLimiterBackend](t *testing.T, mr *miniredis.Miniredis, backend B) {
	l := window_limiter.NewSlidingWindowLimiter[string](backend).
		SetParams(testSlidingWindowLimiterParams)

		// peek the sliding window first
	pv1, err := l.Check(context.Background(), testSlidingWindowLimiterScene, testSlidingWindowLimiterId1)
	require.NoError(t, err)
	require.True(t, pv1.Allow)
	require.Equal(t, 0, pv1.Count)
	require.Equal(t, testSlidingWindowLimiterMaxLimit, pv1.MaxLimit)
	require.NotZero(t, pv1.ExpireAt)

	// take requests
	for i := range testSlidingWindowLimiterMaxLimit {
		v, err := l.Take(context.Background(), testSlidingWindowLimiterScene, testSlidingWindowLimiterId1)
		require.NoError(t, err)
		require.True(t, v.Allow)
		require.Equal(t, i+1, v.Count)
		require.Equal(t, testSlidingWindowLimiterMaxLimit, v.MaxLimit)
		require.NotZero(t, v.ExpireAt)
		time.Sleep(time.Second * 2)
		mr.FastForward(time.Second * 2)
	}
	// full limit, not allowed
	v, err := l.Take(context.Background(), testSlidingWindowLimiterScene, testSlidingWindowLimiterId1)
	require.NoError(t, err)
	require.False(t, v.Allow)
	require.Equal(t, 3, v.Count)
	require.Equal(t, testSlidingWindowLimiterMaxLimit, v.MaxLimit)
	require.NotZero(t, v.ExpireAt)
	// peek the sliding window after full limit
	pv2, err := l.Check(context.Background(), testSlidingWindowLimiterScene, testSlidingWindowLimiterId1)
	require.NoError(t, err)
	require.Equal(t, v, pv2)

	time.Sleep(time.Second * 5)
	mr.FastForward(time.Second * 5)

	// some request out of the window, then allowed
	v, err = l.Take(context.Background(), testSlidingWindowLimiterScene, testSlidingWindowLimiterId1)
	require.NoError(t, err)
	require.True(t, v.Allow)
	require.Equal(t, 3, v.Count)
	require.Equal(t, testSlidingWindowLimiterMaxLimit, v.MaxLimit)
	require.NotZero(t, v.ExpireAt)

	// reset the sliding window
	err = l.Reset(context.Background(), testSlidingWindowLimiterScene, testSlidingWindowLimiterId1)
	require.NoError(t, err)

	// peek the sliding window after reset
	pv3, err := l.Check(context.Background(), testSlidingWindowLimiterScene, testSlidingWindowLimiterId1)
	require.NoError(t, err)
	require.True(t, pv3.Allow)
	require.Equal(t, 0, pv3.Count)
	require.Equal(t, testSlidingWindowLimiterMaxLimit, pv3.MaxLimit)
	require.NotZero(t, pv3.ExpireAt)

	// take requests after reset
	v, err = l.Take(context.Background(), testSlidingWindowLimiterScene, testSlidingWindowLimiterId1)
	require.NoError(t, err)
	require.True(t, v.Allow)
	require.Equal(t, 1, v.Count)
	require.Equal(t, testSlidingWindowLimiterMaxLimit, v.MaxLimit)
	require.NotZero(t, v.ExpireAt)
}

func GenericTest_SlidingWindowLimiter_Lock[B window_limiter.SlidingWindowLimiterBackend](t *testing.T, mr *miniredis.Miniredis, backend B) {
	l := window_limiter.NewSlidingWindowLimiter[string](backend).
		SetParams(testSlidingWindowLimiterParams)

		// take requests
	v, err := l.Take(context.Background(), testSlidingWindowLimiterScene, testSlidingWindowLimiterId1)
	require.NoError(t, err)
	require.True(t, v.Allow)
	require.Equal(t, 1, v.Count)
	require.Equal(t, testSlidingWindowLimiterMaxLimit, v.MaxLimit)
	require.NotZero(t, v.ExpireAt)

	// peek the sliding window after take
	pv1, err := l.Check(context.Background(), testSlidingWindowLimiterScene, testSlidingWindowLimiterId1)
	require.NoError(t, err)
	require.True(t, pv1.Allow)
	require.Equal(t, 1, pv1.Count)
	require.Equal(t, testSlidingWindowLimiterMaxLimit, pv1.MaxLimit)
	require.NotZero(t, pv1.ExpireAt)

	// force lock the sliding window
	v, err = l.Lock(context.Background(), testSlidingWindowLimiterScene, testSlidingWindowLimiterId1)
	require.NoError(t, err)
	require.False(t, v.Allow)
	require.Equal(t, 3, v.Count)
	require.Equal(t, testSlidingWindowLimiterMaxLimit, v.MaxLimit)
	require.NotZero(t, v.ExpireAt)

	// take requests after force lock
	v, err = l.Take(context.Background(), testSlidingWindowLimiterScene, testSlidingWindowLimiterId1)
	require.NoError(t, err)
	require.False(t, v.Allow)
	require.Equal(t, testSlidingWindowLimiterMaxLimit, v.Count)
	require.Equal(t, testSlidingWindowLimiterMaxLimit, v.MaxLimit)
	require.NotZero(t, v.ExpireAt)

	// reset the sliding window
	err = l.Reset(context.Background(), testSlidingWindowLimiterScene, testSlidingWindowLimiterId1)
	require.NoError(t, err)

	// peek the sliding window after reset
	pv2, err := l.Check(context.Background(), testSlidingWindowLimiterScene, testSlidingWindowLimiterId1)
	require.NoError(t, err)
	require.True(t, pv2.Allow)
	require.Equal(t, 0, pv2.Count)
	require.Equal(t, testSlidingWindowLimiterMaxLimit, pv2.MaxLimit)
	require.NotZero(t, pv2.ExpireAt)

	// take requests after reset
	v, err = l.Take(context.Background(), testSlidingWindowLimiterScene, testSlidingWindowLimiterId1)
	require.NoError(t, err)
	require.True(t, v.Allow)
	require.Equal(t, 1, v.Count)
	require.Equal(t, testSlidingWindowLimiterMaxLimit, v.MaxLimit)
	require.NotZero(t, v.ExpireAt)
}
