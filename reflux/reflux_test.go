package reflux

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/thinkgos/proc-extra/testdata"
)

const (
	testPrivFilePath = "../testdata/test.key"
	testPubFilePath  = "../testdata/test.pub"
)

func Test_Reflux_Encrypt_Decrypt(t *testing.T) {
	r, err := New(testdata.PriveKey, testdata.PubKey)
	require.NoError(t, err)
	require.NotNil(t, r.PrivateKey())
	require.NotNil(t, r.PublicKey())

	t.Run("proto", func(t *testing.T) {
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
	})

	t.Run("codec", func(t *testing.T) {
		reg := &testdata.Registration{
			Id:        111,
			OpenId:    "222",
			ExpiredAt: time.Now().Unix(),
			Code:      "abcdefg",
		}

		tk, err := r.Encrypt(reg)
		require.NoError(t, err)

		got := &testdata.Registration{}
		err = r.Decrypt(tk, got)
		require.NoError(t, err)

		require.Equal(t, reg, got)
	})
}

func Test_Reflux_Sign_Verify(t *testing.T) {
	r, err := New(testdata.PriveKey, testdata.PubKey)
	require.NoError(t, err)

	t.Run("proto", func(t *testing.T) {
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
	})
	t.Run("codec", func(t *testing.T) {
		reg := &testdata.Registration{
			Id:        111,
			OpenId:    "222",
			ExpiredAt: time.Now().Unix(),
			Code:      "abcdefg",
		}

		tk, err := r.Sign(reg)
		require.NoError(t, err)
		t.Log(len(tk))

		err = r.Verify(tk, reg)
		require.NoError(t, err)
	})
}

func Test_Reflux_Encrypt_Decrypt_Use_File_Codec(t *testing.T) {
	r, err := New(testPrivFilePath, testPubFilePath, WithCodecString(base64.RawURLEncoding), WithCodec(new(CodecJSON)))
	require.NoError(t, err)

	t.Run("proto", func(t *testing.T) {
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
	})
	t.Run("codec", func(t *testing.T) {
		reg := &testdata.Registration{
			Id:        111,
			OpenId:    "222",
			ExpiredAt: time.Now().Unix(),
			Code:      "abcdefg",
		}

		tk, err := r.Encrypt(reg)
		require.NoError(t, err)
		t.Log(len(tk))

		got := &testdata.Registration{}
		err = r.Decrypt(tk, got)
		require.NoError(t, err)

		require.Equal(t, reg, got)
	})
}

func Test_Reflux_Sign_Verify_Use_File_Codec(t *testing.T) {
	r, err := New(testPrivFilePath, testPubFilePath, WithCodecString(base64.RawURLEncoding))
	require.NoError(t, err)

	t.Run("proto", func(t *testing.T) {
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
	})
	t.Run("codec", func(t *testing.T) {
		reg := &testdata.Registration{
			Id:        111,
			OpenId:    "222",
			ExpiredAt: time.Now().Unix(),
			Code:      "abcdefg",
		}

		tk, err := r.Sign(reg)
		require.NoError(t, err)
		t.Log(len(tk))

		err = r.Verify(tk, reg)
		require.NoError(t, err)
	})
}
