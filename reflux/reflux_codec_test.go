package reflux

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/thinkgos/proc-extra/testdata"
)

func Test_Codec_Encrypt_Decrypt(t *testing.T) {
	r, err := New(testdata.PriveKey, testdata.PubKey)
	require.NoError(t, err)
	require.NotNil(t, r.PrivateKey())
	require.NotNil(t, r.PublicKey())

	reg := &testJson{
		Id:        111,
		OpenId:    "222",
		ExpiredAt: time.Now().Unix(),
		Code:      "abcdefg",
	}

	tk, err := r.EncryptCodec(reg)
	require.NoError(t, err)

	got := &testJson{}
	err = r.DecryptCodec(tk, got)
	require.NoError(t, err)

	require.Equal(t, reg, got)
}

func Test_Codec_Sign_Verify(t *testing.T) {
	r, err := New(testdata.PriveKey, testdata.PubKey)
	require.NoError(t, err)

	reg := &testJson{
		Id:        111,
		OpenId:    "222",
		ExpiredAt: time.Now().Unix(),
		Code:      "abcdefg",
	}

	tk, err := r.SignCodec(reg)
	require.NoError(t, err)
	t.Log(len(tk))

	err = r.VerifyCodec(tk, reg)
	require.NoError(t, err)
}

func Test_Codec_Encrypt_Decrypt_Use_File_CodecJSON(t *testing.T) {
	r, err := New(testPrivFilePath, testPubFilePath, WithCodecString(base64.RawURLEncoding), WithCodec(new(CodecJSON)))
	require.NoError(t, err)

	reg := &testJson{
		Id:        111,
		OpenId:    "222",
		ExpiredAt: time.Now().Unix(),
		Code:      "abcdefg",
	}

	tk, err := r.EncryptCodec(reg)
	require.NoError(t, err)
	t.Log(len(tk))

	got := &testJson{}
	err = r.DecryptCodec(tk, got)
	require.NoError(t, err)

	require.Equal(t, reg, got)
}

func Test_Codec_Sign_Verify_Use_File_Codec(t *testing.T) {
	r, err := New(testPrivFilePath, testPubFilePath, WithCodecString(base64.RawURLEncoding))
	require.NoError(t, err)

	reg := &testJson{
		Id:        111,
		OpenId:    "222",
		ExpiredAt: time.Now().Unix(),
		Code:      "abcdefg",
	}

	tk, err := r.SignCodec(reg)
	require.NoError(t, err)
	t.Log(len(tk))

	err = r.VerifyCodec(tk, reg)
	require.NoError(t, err)
}
