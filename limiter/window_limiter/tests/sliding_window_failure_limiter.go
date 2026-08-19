package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"
	"github.com/thinkgos/proc-extra/limiter/window_limiter"
)

const testSlidingWindowFailureLimiterScene testSceneType = "test_kind"

const (
	testSlidingWindowFailureLimiterKeyPrefix  = "period:failure:limiter:test-scene:"
	testSlidingWindowFailureLimiterWindow     = 10
	testSlidingWindowFailureLimiterMaxAttempt = 3
)

const testSlidingWindowFailureLimiterId1 = "id1"

var errTestSlidingWindowFailureLimiter = errors.New("a test error")

var testSlidingWindowFailureLimiterParam = &window_limiter.SlidingWindowLimiterParam{
	Window:   testSlidingWindowFailureLimiterWindow,
	MaxLimit: testSlidingWindowFailureLimiterMaxAttempt,
}

func GenericTest_SlidingWindowFailureLimiter_Work[B window_limiter.SlidingWindowFailureLimiterBackend](t *testing.T, mr *miniredis.Miniredis, backend B) {
	l := window_limiter.NewSlidingWindowFailureLimiter[testSceneType](backend).
		SetKeyPrefix(testSlidingWindowFailureLimiterKeyPrefix).
		SetParam(testSlidingWindowFailureLimiterParam)

	// check the sliding window first
	pv1, err := l.Check(context.Background(), testSlidingWindowFailureLimiterScene, testSlidingWindowFailureLimiterId1)
	require.NoError(t, err)
	require.True(t, pv1.Allow)
	require.Equal(t, 0, pv1.Failures)
	require.Equal(t, testSlidingWindowFailureLimiterMaxAttempt, pv1.MaxFailures)
	require.NotZero(t, pv1.ExpireAt)

	// evaluate error operations
	for i := range testSlidingWindowFailureLimiterMaxAttempt {
		v, err := l.EvaluateErr(context.Background(), testSlidingWindowFailureLimiterScene, testSlidingWindowFailureLimiterId1, errTestSlidingWindowFailureLimiter)
		require.NoError(t, err)
		require.True(t, v.Allow)
		require.Equal(t, i+1, v.Failures)
		require.Equal(t, testSlidingWindowFailureLimiterMaxAttempt, v.MaxFailures)
		require.NotZero(t, v.ExpireAt)
		time.Sleep(time.Second * 2)
		mr.FastForward(time.Second * 2)
	}
	// full limit, not allowed
	v, err := l.EvaluateErr(context.Background(), testSlidingWindowFailureLimiterScene, testSlidingWindowFailureLimiterId1, nil)
	require.NoError(t, err)
	require.False(t, v.Allow)
	require.Equal(t, 3, v.Failures)
	require.Equal(t, testSlidingWindowFailureLimiterMaxAttempt, v.MaxFailures)
	require.NotZero(t, v.ExpireAt)
	// check the sliding window after full limit
	pv2, err := l.Check(context.Background(), testSlidingWindowFailureLimiterScene, testSlidingWindowFailureLimiterId1)
	require.NoError(t, err)
	require.Equal(t, v, pv2)

	time.Sleep(time.Second * 5)
	mr.FastForward(time.Second * 5)

	// some request out of the window, then allowed
	v, err = l.EvaluateErr(context.Background(), testSlidingWindowFailureLimiterScene, testSlidingWindowFailureLimiterId1, nil)
	require.NoError(t, err)
	require.True(t, v.Allow)
	require.Equal(t, 0, v.Failures)
	require.Equal(t, testSlidingWindowFailureLimiterMaxAttempt, v.MaxFailures)
	require.NotZero(t, v.ExpireAt)

	// check the sliding window after success
	pv3, err := l.Check(context.Background(), testSlidingWindowFailureLimiterScene, testSlidingWindowFailureLimiterId1)
	require.NoError(t, err)
	require.True(t, pv3.Allow)
	require.Equal(t, 0, pv3.Failures)
	require.Equal(t, testSlidingWindowFailureLimiterMaxAttempt, pv3.MaxFailures)
	require.NotZero(t, pv3.ExpireAt)

	// evaluate requests operation after success
	v, err = l.EvaluateErr(context.Background(), testSlidingWindowFailureLimiterScene, testSlidingWindowFailureLimiterId1, errTestSlidingWindowFailureLimiter)
	require.NoError(t, err)
	require.True(t, v.Allow)
	require.Equal(t, 1, v.Failures)
	require.Equal(t, testSlidingWindowFailureLimiterMaxAttempt, v.MaxFailures)
	require.NotZero(t, v.ExpireAt)
}

func GenericTest_SlidingWindowFailureLimiter_Lock[B window_limiter.SlidingWindowFailureLimiterBackend](t *testing.T, mr *miniredis.Miniredis, backend B) {
	l := window_limiter.NewSlidingWindowFailureLimiter[testSceneType](backend).
		SetKeyPrefix(testSlidingWindowFailureLimiterKeyPrefix).
		SetParam(testSlidingWindowFailureLimiterParam)

	// evaluate requests
	v, err := l.EvaluateErr(context.Background(), testSlidingWindowFailureLimiterScene, testSlidingWindowFailureLimiterId1, errTestSlidingWindowFailureLimiter)
	require.NoError(t, err)
	require.True(t, v.Allow)
	require.Equal(t, 1, v.Failures)
	require.Equal(t, testSlidingWindowFailureLimiterMaxAttempt, v.MaxFailures)
	require.NotZero(t, v.ExpireAt)

	// check the sliding window after evaluate
	pv1, err := l.Check(context.Background(), testSlidingWindowFailureLimiterScene, testSlidingWindowFailureLimiterId1)
	require.NoError(t, err)
	require.True(t, pv1.Allow)
	require.Equal(t, 1, pv1.Failures)
	require.Equal(t, testSlidingWindowFailureLimiterMaxAttempt, pv1.MaxFailures)
	require.NotZero(t, pv1.ExpireAt)

	// lock the sliding window
	v, err = l.Lock(context.Background(), testSlidingWindowFailureLimiterScene, testSlidingWindowFailureLimiterId1)
	require.NoError(t, err)
	require.False(t, v.Allow)
	require.Equal(t, 3, v.Failures)
	require.Equal(t, testSlidingWindowFailureLimiterMaxAttempt, v.MaxFailures)
	require.NotZero(t, v.ExpireAt)

	// evaluate requests operation after lock
	v, err = l.EvaluateErr(context.Background(), testSlidingWindowFailureLimiterScene, testSlidingWindowFailureLimiterId1, errTestSlidingWindowFailureLimiter)
	require.NoError(t, err)
	require.False(t, v.Allow)
	require.Equal(t, testSlidingWindowFailureLimiterMaxAttempt, v.Failures)
	require.Equal(t, testSlidingWindowFailureLimiterMaxAttempt, v.MaxFailures)
	require.NotZero(t, v.ExpireAt)

	// reset the sliding window
	err = l.Reset(context.Background(), testSlidingWindowFailureLimiterScene, testSlidingWindowFailureLimiterId1)
	require.NoError(t, err)

	// check the sliding window after reset
	pv2, err := l.Check(context.Background(), testSlidingWindowFailureLimiterScene, testSlidingWindowFailureLimiterId1)
	require.NoError(t, err)
	require.True(t, pv2.Allow)
	require.Equal(t, 0, pv2.Failures)
	require.Equal(t, testSlidingWindowFailureLimiterMaxAttempt, pv2.MaxFailures)
	require.NotZero(t, pv2.ExpireAt)

	// evaluate requests operation after reset
	v, err = l.EvaluateErr(context.Background(), testSlidingWindowFailureLimiterScene, testSlidingWindowFailureLimiterId1, errTestSlidingWindowFailureLimiter)
	require.NoError(t, err)
	require.True(t, v.Allow)
	require.Equal(t, 1, v.Failures)
	require.Equal(t, testSlidingWindowFailureLimiterMaxAttempt, v.MaxFailures)
	require.NotZero(t, v.ExpireAt)
}
