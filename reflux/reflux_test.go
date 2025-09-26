package reflux

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thinkgos/proc-extra/cert"
)

const (
	testPrivFilePath = "../testdata/test.key"
	testPubFilePath  = "../testdata/test.pub"
)

type testJson struct {
	Id        int64  `json:"id"`
	OpenId    string `json:"openId"`
	ExpiredAt int64  `json:"expiredAt"`
	Code      string `json:"code"`
}

func Test_Encrypt_Decrypt(t *testing.T) {
	priKey, err := cert.LoadRSAPrivateKeyFromFile(testPrivFilePath)
	require.NoError(t, err)
	pubKey, err := cert.LoadRSAPublicKeyFromPemFile(testPubFilePath)
	require.NoError(t, err)

	want := []byte("helloworld,this is golang language. welcome")

	mid, err := Encrypt(pubKey, want)
	require.NoError(t, err)

	got, err := Decrypt(priKey, mid)
	require.NoError(t, err)

	require.Equal(t, want, got)
}
