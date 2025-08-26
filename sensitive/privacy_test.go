package sensitive

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Privacy(t *testing.T) {
	p := New(
		WithIvGen(IvGenTimestamp),
		WithIvChecker(IvCheckerTimestamp(time.Minute*5)),
	)
	secret := []byte("0123456789abcdef")

	rawText := []byte("hhh666")
	cipherText, err := p.Encrypt(secret, rawText)
	require.NoError(t, err)
	wantText, err := p.Decrypt(secret, cipherText)
	require.NoError(t, err)
	require.Equal(t, wantText, rawText)
	t.Logf("raw: %v, cipher: %v\n", string(rawText), cipherText)

	rawText = []byte("hello world")
	cipherText, err = p.Encrypt(secret, rawText)
	require.NoError(t, err)
	wantText, err = p.Decrypt(secret, cipherText)
	require.NoError(t, err)
	require.Equal(t, wantText, rawText)
	t.Logf("raw: %v, cipher: %v\n", string(rawText), cipherText)
}
