package tests

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"

	"github.com/thinkgos/proc-extra/limiter/limit_verified"
)

const (
	testKeyPrefix      = "limit:verifier:scene:"
	testSceneInvalid   = "test_scene_invalid"
	testSceneNormal    = "test_scene1"
	testSceneOverQuota = "test_scene2"
	testSceneTierLimit = "test_scene3"
	target             = "112233"
	code               = "123456"
	badCode            = "654321"
)

var testParams = map[string]*limit_verified.Param{
	testSceneNormal: limit_verified.NewParam(testSceneNormal),
	testSceneOverQuota: {
		Scene:           testSceneOverQuota,
		Window:          time.Hour * 24,
		Quota:           1,
		CodeExpires:     300,
		CodeMaxAttempts: 3,
	},
	testSceneTierLimit: {
		Scene:  testSceneTierLimit,
		Window: time.Hour * 24,
		Quota:  30,
		WindowTiers: []limit_verified.WindowTier{
			{Window: time.Minute, Quota: 1},
		},
		CodeExpires:     300,
		CodeMaxAttempts: 3,
	},
}

type TestProvider struct{}

func (t TestProvider) Name() string                                            { return "test_provider1" }
func (t TestProvider) SendCode(ctx context.Context, target, code string) error { return nil }

type TestErrProvider struct{}

func (t TestErrProvider) Name() string { return "test_provider2" }
func (t TestErrProvider) SendCode(ctx context.Context, target, code string) error {
	return errors.New("发送失败")
}

func GenericTest_Name[B limit_verified.LimitVerifiedBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := limit_verified.NewLimitVerifiedRegistry[string](new(TestProvider), backend).SetKeyPrefix(testKeyPrefix)
	require.Equal(t, "test_provider1", l.Name())
}

func GenericTest_InvalidScene[B limit_verified.LimitVerifiedBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := limit_verified.NewLimitVerifiedRegistry[string](new(TestProvider), backend)

	_, err := l.SendCode(context.Background(), testSceneInvalid, target, code)
	require.ErrorIs(t, err, limit_verified.ErrSceneParamNotFound)
	_, err = l.VerifyCode(context.Background(), testSceneInvalid, target, code)
	require.ErrorIs(t, err, limit_verified.ErrSceneParamNotFound)
}

func GenericTest_Work[B limit_verified.LimitVerifiedBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := limit_verified.NewLimitVerifiedRegistry[string](new(TestProvider), backend).RegistersParams(testParams)

	result, err := l.SendCode(context.Background(), testSceneNormal, target, code)
	require.NoError(t, err)
	require.Equal(t, limit_verified.EvaluateStatus_Success, result.Status)

	vr, err := l.VerifyCode(context.Background(), testSceneNormal, target, code)
	require.NoError(t, err)
	require.Equal(t, limit_verified.VerifyStatus_Success, vr.Status)
}

func GenericTest_SendCode_Failure[B limit_verified.LimitVerifiedBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := limit_verified.NewLimitVerifiedRegistry[string](new(TestErrProvider), backend).RegistersParams(testParams)

	_, err := l.SendCode(context.Background(), testSceneNormal, target, code)
	require.Error(t, err)
}

func GenericTest_SendCode_OverQuota[B limit_verified.LimitVerifiedBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	var success uint32
	var failed uint32

	l := limit_verified.NewLimitVerifiedRegistry[string](new(TestProvider), backend).RegistersParams(testParams)

	wg := &sync.WaitGroup{}
	wg.Add(15)
	for range 15 {
		go func() {
			defer wg.Done()

			result, err := l.SendCode(context.Background(), testSceneOverQuota, target, code)
			require.NoError(t, err)
			switch result.Status {
			case limit_verified.EvaluateStatus_Success:
				atomic.AddUint32(&success, 1)
			case limit_verified.EvaluateStatus_OverQuota:
				atomic.AddUint32(&failed, 1)
			case limit_verified.EvaluateStatus_TooFrequently:
				fallthrough
			default:
				require.Fail(t, "unexpected evaluate code")
			}
		}()
	}

	wg.Wait()
	require.Equal(t, uint32(1), success)
	require.Equal(t, uint32(14), failed)
}

func GenericTest_SendCode_ResendTooFrequently[B limit_verified.LimitVerifiedBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	var success uint32
	var failed uint32

	l := limit_verified.NewLimitVerifiedRegistry[string](new(TestProvider), backend).RegistersParams(testParams)

	wg := &sync.WaitGroup{}
	wg.Add(15)
	for range 15 {
		go func() {
			defer wg.Done()

			result, err := l.SendCode(context.Background(), testSceneTierLimit, target, code)
			require.NoError(t, err)
			switch result.Status {
			case limit_verified.EvaluateStatus_Success:
				atomic.AddUint32(&success, 1)
			case limit_verified.EvaluateStatus_TooFrequently:
				atomic.AddUint32(&failed, 1)
			case limit_verified.EvaluateStatus_OverQuota:
				fallthrough
			default:
				require.Fail(t, "unexpected evaluate code")
			}
		}()
	}

	wg.Wait()
	require.Equal(t, uint32(1), success)
	require.Equal(t, uint32(14), failed)
}

func GenericTest_VerifyCode_Expired[B limit_verified.LimitVerifiedBackend](t *testing.T, mr *miniredis.Miniredis, backend B) {
	l := limit_verified.NewLimitVerifiedRegistry[string](new(TestProvider), backend).RegistersParams(testParams)

	// 没有验证码
	vr, err := l.VerifyCode(context.Background(), testSceneNormal, target, code)
	require.NoError(t, err)
	require.Equal(t, limit_verified.VerifyStatus_Expired, vr.Status)

	// 验证码过期
	result, err := l.SendCode(context.Background(), testSceneNormal, target, code)
	require.NoError(t, err)
	require.Equal(t, limit_verified.EvaluateStatus_Success, result.Status)

	mr.FastForward(time.Second * 301)
	vr, err = l.VerifyCode(context.Background(), testSceneNormal, target, code)
	require.NoError(t, err)
	require.Equal(t, limit_verified.VerifyStatus_Expired, vr.Status)
}

func GenericTest_SendCode_Rollback[B limit_verified.LimitVerifiedBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	// 先用失败 provider 发送, 触发 rollback
	l1 := limit_verified.NewLimitVerifiedRegistry[string](new(TestErrProvider), backend).RegistersParams(testParams)
	_, err := l1.SendCode(context.Background(), testSceneOverQuota, target, code)
	require.Error(t, err)

	// 再用成功 provider 发送, 如果 rollback 正常, 配额应已恢复, 不会 OverQuota
	l2 := limit_verified.NewLimitVerifiedRegistry[string](new(TestProvider), backend).RegistersParams(testParams)
	result, err := l2.SendCode(context.Background(), testSceneOverQuota, target, code)
	require.NoError(t, err)
	require.Equal(t, limit_verified.EvaluateStatus_Success, result.Status)
}

func GenericTest_VerifyCode_ReachMaxAttempt[B limit_verified.LimitVerifiedBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	var failedExpired uint32
	var failedVerify uint32

	l := limit_verified.NewLimitVerifiedRegistry[string](new(TestProvider), backend).RegistersParams(testParams)

	result, err := l.SendCode(context.Background(), testSceneNormal, target, code)
	require.NoError(t, err)
	require.Equal(t, limit_verified.EvaluateStatus_Success, result.Status)

	wg := &sync.WaitGroup{}
	wg.Add(15)
	for range 15 {
		go func() {
			defer wg.Done()

			vr, err := l.VerifyCode(context.Background(), testSceneNormal, target, badCode)
			require.NoError(t, err)
			switch vr.Status {
			case limit_verified.VerifyStatus_Failure:
				atomic.AddUint32(&failedVerify, 1)
			case limit_verified.VerifyStatus_Expired:
				atomic.AddUint32(&failedExpired, 1)
			case limit_verified.VerifyStatus_Success:
				fallthrough
			default:
				require.Fail(t, "unexpected verify code")
			}
		}()
	}
	wg.Wait()
	require.Equal(t, uint32(3), failedVerify)
	require.Equal(t, uint32(12), failedExpired)
}
