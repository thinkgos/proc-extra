package reflux

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/thinkgos/proc-extra/testdata"
	"google.golang.org/protobuf/proto"
)

func Test_Proto_Encrypt_Decrypt(t *testing.T) {
	r, err := New(testdata.PriveKey, testdata.PubKey)
	require.NoError(t, err)
	require.NotNil(t, r.PrivateKey())
	require.NotNil(t, r.PublicKey())

	reg := &testdata.Registration{
		Id:        111,
		OpenId:    "222",
		ExpiredAt: time.Now().Unix(),
		Code:      "abcdefg",
	}

	tk, err := r.EncryptProto(reg)
	require.NoError(t, err)
	t.Log(len(tk))

	got := &testdata.Registration{}
	err = r.DecryptProto(tk, got)
	require.NoError(t, err)

	require.True(t, proto.Equal(reg, got))
}

func Test_Proto_Sign_Verify(t *testing.T) {
	r, err := New(testdata.PriveKey, testdata.PubKey)
	require.NoError(t, err)

	reg := &testdata.Registration{
		Id:        111,
		OpenId:    "222",
		ExpiredAt: time.Now().Unix(),
		Code:      "abcdefg",
	}

	tk, err := r.SignProto(reg)
	require.NoError(t, err)
	t.Log(len(tk))

	err = r.VerifyProto(tk, reg)
	require.NoError(t, err)
}

func Test_Proto_Encrypt_Decrypt_Use_File_CodecJSON(t *testing.T) {
	r, err := New(testPrivFilePath, testPubFilePath, WithCodecString(base64.RawURLEncoding), WithCodec(new(CodecJSON)))
	require.NoError(t, err)

	reg := &testdata.Registration{
		Id:        111,
		OpenId:    "222",
		ExpiredAt: time.Now().Unix(),
		Code:      "abcdefg",
	}

	tk, err := r.EncryptProto(reg)
	require.NoError(t, err)
	t.Log(len(tk))

	got := &testdata.Registration{}
	err = r.DecryptProto(tk, got)
	require.NoError(t, err)

	require.True(t, proto.Equal(reg, got))
}

func Test_Proto_Sign_Verify_Use_File_Codec(t *testing.T) {
	r, err := New(testPrivFilePath, testPubFilePath, WithCodecString(base64.RawURLEncoding))
	require.NoError(t, err)

	reg := &testdata.Registration{
		Id:        111,
		OpenId:    "222",
		ExpiredAt: time.Now().Unix(),
		Code:      "abcdefg",
	}

	tk, err := r.SignProto(reg)
	require.NoError(t, err)
	t.Log(len(tk))

	err = r.VerifyProto(tk, reg)
	require.NoError(t, err)
}
