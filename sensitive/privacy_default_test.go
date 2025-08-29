package sensitive

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_PrivacyDefault(t *testing.T) {
	secret := []byte("0123456789abcdef")

	rawText := []byte("hhh666")
	cipherText, err := EncryptViaRandom(secret, rawText)
	require.NoError(t, err)
	wantText, err := DecryptViaRandom(secret, cipherText)
	require.NoError(t, err)
	require.Equal(t, wantText, rawText)
	t.Logf("raw: %v, cipher: %v\n", string(rawText), cipherText)

	rawText = []byte("hello world")
	cipherText, err = EncryptViaTimestamp(secret, rawText)
	require.NoError(t, err)
	wantText, err = DecryptViaTimestamp(secret, cipherText)
	require.NoError(t, err)
	require.Equal(t, wantText, rawText)
	t.Logf("raw: %v, cipher: %v\n", string(rawText), cipherText)
}
