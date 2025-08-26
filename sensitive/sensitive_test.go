package sensitive

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

var _ EncryptSensitiveAble = (*SensitiveObject)(nil)
var _ DecryptSensitiveAble = (*SensitiveObject)(nil)

type SensitiveObject struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// 实现 DesensitizeAble[T any] 接口
func (p *SensitiveObject) EncryptSensitive(secret []byte) (string, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// 实现 DesensitizeAble[T any] 接口
func (p *SensitiveObject) DecryptSensitive(secret []byte, cipherText string) error {
	b, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, p)
}

func Test_Sensitive(t *testing.T) {
	secret := []byte("0123456789abcdef")
	want := &SensitiveObject{Name: "Alice", Age: 26}
	// 加密
	cipherText, err := want.EncryptSensitive(secret)
	require.NoError(t, err)

	// 解密
	got := &SensitiveObject{}
	err = got.DecryptSensitive(secret, cipherText)
	require.NoError(t, err)
	require.Equal(t, want, got)
}
