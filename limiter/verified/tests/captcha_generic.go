package tests

import (
	"context"
	"math/bits"
	"math/rand"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"

	"github.com/thinkgos/proc-extra/limiter/verified"
)

const (
	testDriverName        = "test_driver_name"
	unsupportedDriverName = "unsupported_driver_name"
)

type testSceneType string

func (s testSceneType) Value() string { return string(s) }

const testScene testSceneType = "test_kind"
const (
	testKeyPrefix     = "verified:captcha:test-scene:"
	testKeyExpires    = time.Minute * 5
	testMaxAttempts_1 = 1
	testMaxAttempts_3 = 3
)

const (
	question    = "1+1"
	rightAnswer = "2"
	wrongAnswer = "3"
)

var testCaptchaSceneParam = &verified.Param{
	KeyExpires:  testKeyExpires,
	MaxAttempts: testMaxAttempts_1,
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

var _ verified.CaptchaDriver = (*TestCaptchaDriver)(nil)

type TestCaptchaDriver struct{}

func (t TestCaptchaDriver) Driver(dName string) verified.ChallengeProvider {
	if dName == unsupportedDriverName {
		return new(verified.UnsupportedChallengeProvider)
	}
	return new(TestChallenge)
}

type TestChallenge struct{}

func (t TestChallenge) Name() string { return testDriverName }
func (t TestChallenge) GenerateChallenge(ctx context.Context) (*verified.Challenge, error) {
	return &verified.Challenge{
		Id:       randString(6),
		Question: question,
		Answer:   rightAnswer,
	}, nil
}

func GenericTest_Captcha_UnsupportedChallengeProvider[B verified.StorageBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := verified.NewCaptcha[testSceneType](new(TestCaptchaDriver), backend).
		SetKeyPrefix(testKeyPrefix).
		SetSceneParam(testScene, testCaptchaSceneParam)

	_, _, err := l.Generate(context.Background(), unsupportedDriverName, testScene)
	require.Error(t, err)
}

func GenericTest_Captcha_InMaxAttempts[B verified.StorageBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := verified.NewCaptcha[testSceneType](new(TestCaptchaDriver), backend).
		SetKeyPrefix(testKeyPrefix).
		SetParam(&verified.Param{
			KeyExpires:  testKeyExpires,
			MaxAttempts: testMaxAttempts_3,
		})

	id, _, err := l.Generate(context.Background(), testDriverName, testScene)
	require.NoError(t, err)

	b, err := l.Verify(context.Background(), testScene, id, wrongAnswer)
	require.NoError(t, err)
	require.False(t, b)
	b, err = l.Verify(context.Background(), testScene, id, wrongAnswer)
	require.NoError(t, err)
	require.False(t, b)
	b, err = l.Verify(context.Background(), testScene, id, rightAnswer)
	require.NoError(t, err)
	require.True(t, b)
}

func GenericTest_Captcha_OverMaxAttempts[B verified.StorageBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := verified.NewCaptcha[testSceneType](new(TestCaptchaDriver), backend).
		SetKeyPrefix(testKeyPrefix).
		SetParam(&verified.Param{
			KeyExpires:  testKeyExpires,
			MaxAttempts: 6,
		})

	id, _, err := l.Generate(context.Background(), testDriverName, testScene)
	require.NoError(t, err)

	for range 6 {
		b, err := l.Verify(context.Background(), testScene, id, wrongAnswer)
		require.NoError(t, err)
		require.False(t, b)
	}
	b, err := l.Verify(context.Background(), testScene, id, rightAnswer)
	require.NoError(t, err)
	require.False(t, b)
}

func GenericTest_Captcha_OneShot[B verified.StorageBackend](t *testing.T, _ *miniredis.Miniredis, backend B) {
	l := verified.NewCaptcha[testSceneType](new(TestCaptchaDriver), backend).
		SetKeyPrefix(testKeyPrefix).
		SetSceneParam(testScene, testCaptchaSceneParam)

	id, _, err := l.Generate(context.Background(), testDriverName, testScene)
	require.NoError(t, err)

	b, err := l.Verify(context.Background(), testScene, id, rightAnswer)
	require.NoError(t, err)
	require.True(t, b)

	b, err = l.Verify(context.Background(), testScene, id, rightAnswer)
	require.NoError(t, err)
	require.False(t, b)
}

func GenericTest_Captcha_OneShot_Timeout[B verified.StorageBackend](t *testing.T, mr *miniredis.Miniredis, backend B) {
	l := verified.NewCaptcha[testSceneType](new(TestCaptchaDriver), backend).
		SetKeyPrefix(testKeyPrefix).
		SetSceneParam(testScene, &verified.Param{
			KeyExpires:  time.Second * 1,
			MaxAttempts: testMaxAttempts_1,
		})

	id, _, err := l.Generate(context.Background(), testDriverName, testScene)
	require.NoError(t, err)

	mr.FastForward(time.Second * 2) // time.Sleep(time.Second * 2)

	b, err := l.Verify(context.Background(), testScene, id, rightAnswer)
	require.NoError(t, err)
	require.False(t, b)
}
