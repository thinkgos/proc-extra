package tests

import (
	"context"
	"math/bits"
	"math/rand"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/thinkgos/proc-extra/limiter/verified"
)

const (
	testDriverName        = "test_driver_name"
	unsupportedDriverName = "unsupported_driver_name"
)

const testInvalidKind = "invalid_kind"
const testKind = "test_kind"
const (
	testKeyPrefix     = "verified:captcha:test-kind:"
	testKeyExpires    = time.Minute * 5
	testMaxAttempts_1 = 1
)

const (
	question    = "1+1"
	rightAnswer = "2"
	wrongAnswer = "3"
)

var testCaptchaParams = map[string]*verified.Param{
	testKind: {
		KeyPrefix:   testKeyPrefix,
		KeyExpires:  testKeyExpires,
		MaxAttempts: testMaxAttempts_1,
	},
}

var defaultAlphabet = []byte("QWERTYUIOPLKJHGFDSAZXCVBNMabcdefghijklmnopqrstuvwxyz")

func randString(length int) string {
	b := make([]byte, length)
	bn := bits.Len(uint(len(defaultAlphabet)))
	mask := int64(1)<<bn - 1
	max := 63 / bn
	r := rand.New(rand.NewSource(time.Now().UnixNano() + rand.Int63() + rand.Int63()))

	// A rand.Int63() generates 63 random bits, enough for alphabets letters!
	for i, cache, remain := 0, r.Int63(), max; i < length; {
		if remain == 0 {
			cache, remain = r.Int63(), max
		}
		if idx := int(cache & mask); idx < len(defaultAlphabet) {
			b[i] = defaultAlphabet[idx]
			i++
		}
		cache >>= bn
		remain--
	}
	return string(b)
}

var _ verified.CaptchaProvider = (*TestCaptchaProvider)(nil)

type TestCaptchaProvider struct{}

func (t TestCaptchaProvider) AcquireDriver(dName string) verified.CaptchaDriver {
	if dName == unsupportedDriverName {
		return new(verified.UnsupportedCaptchaDriver)
	}
	return new(TestCaptchaDriver)
}

type TestCaptchaDriver struct{}

func (t TestCaptchaDriver) Name() string { return testDriverName }
func (t TestCaptchaDriver) GenerateQuestionAnswer() (*verified.Challenge, error) {
	return &verified.Challenge{
		Id:       randString(6),
		Question: question,
		Answer:   rightAnswer,
	}, nil
}

func GenericTestCaptcha_Improve_Cover[B verified.StorageBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := verified.NewCaptchaVerified[string](new(TestCaptchaProvider), backend).
		SetParams(testCaptchaParams)
	require.Equal(t, testDriverName, l.Name(testDriverName))

	id, _, err := l.Generate(context.Background(), testDriverName, testInvalidKind)
	require.ErrorIs(t, verified.ErrParamKindNotFound, err)

	_, err = l.Verify(context.Background(), testInvalidKind, id, rightAnswer)
	require.ErrorIs(t, verified.ErrParamKindNotFound, err)
}

func GenericTestCaptcha_Unsupported_Driver[B verified.StorageBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := verified.NewCaptchaVerified[string](new(TestCaptchaProvider), backend).
		SetParams(testCaptchaParams)

	_, _, err := l.Generate(context.Background(), unsupportedDriverName, testKind)
	assert.Error(t, err)
}

func GenericTestCaptcha_OneTime[B verified.StorageBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := verified.NewCaptchaVerified[string](new(TestCaptchaProvider), backend).
		SetParams(testCaptchaParams)

	id, _, err := l.Generate(context.Background(), testDriverName, testKind)
	assert.NoError(t, err)

	b, err := l.Verify(context.Background(), testKind, id, rightAnswer)
	require.NoError(t, err)
	require.True(t, b)

	b, err = l.Verify(context.Background(), testKind, id, rightAnswer)
	require.NoError(t, err)
	require.False(t, b)
}

func GenericTestCaptcha_In_MaxAttempts[B verified.StorageBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := verified.NewCaptchaVerified[string](new(TestCaptchaProvider), backend).
		SetParams(testCaptchaParams)

	id, _, err := l.Generate(
		context.Background(),
		testDriverName,
		testKind,
		verified.WithKeyExpires(time.Minute*5),
		verified.WithMaxAttempts(3),
	)
	assert.NoError(t, err)

	b, err := l.Verify(context.Background(), testKind, id, wrongAnswer)
	require.NoError(t, err)
	require.False(t, b)
	b, err = l.Verify(context.Background(), testKind, id, wrongAnswer)
	require.NoError(t, err)
	require.False(t, b)
	b, err = l.Verify(context.Background(), testKind, id, rightAnswer)
	require.NoError(t, err)
	require.True(t, b)
}

func GenericTestCaptcha_Over_MaxAttempts[B verified.StorageBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := verified.NewCaptchaVerified[string](new(TestCaptchaProvider), backend).
		SetParams(testCaptchaParams)

	id, _, err := l.Generate(context.Background(), testDriverName, testKind,
		verified.WithKeyExpires(time.Minute*5),
		verified.WithMaxAttempts(6),
	)
	assert.NoError(t, err)

	for range 6 {
		b, err := l.Verify(context.Background(), testKind, id, wrongAnswer)
		require.NoError(t, err)
		require.False(t, b)
	}
	b, err := l.Verify(context.Background(), testKind, id, rightAnswer)
	require.NoError(t, err)
	require.False(t, b)
}

func GenericTestCaptcha_Onetime_Timeout[B verified.StorageBackend](t *testing.T, mr *miniredis.Miniredis, backend B) {
	l := verified.NewCaptchaVerified[string](new(TestCaptchaProvider), backend).
		SetParams(testCaptchaParams)

	id, _, err := l.Generate(context.Background(), testDriverName, testKind, verified.WithKeyExpires(time.Second*1))
	assert.NoError(t, err)

	mr.FastForward(time.Second * 2) // time.Sleep(time.Second * 2)

	b, err := l.Verify(context.Background(), testKind, id, rightAnswer)
	require.NoError(t, err)
	require.False(t, b)
}
