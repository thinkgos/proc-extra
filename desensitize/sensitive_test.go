package desensitize

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

var _ SensitiveEncryptAble = (*SensitiveObject)(nil)
var _ SensitiveDecryptAble = (*SensitiveObject)(nil)

type SensitiveObject struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// 实现 DesensitizeAble[T any] 接口
func (p *SensitiveObject) IntoSensitive(identityId string) error {
	p.Name = base64.StdEncoding.EncodeToString([]byte(p.Name))
	return nil
}

// 实现 DesensitizeAble[T any] 接口
func (p *SensitiveObject) FromSensitive(identityId string) error {
	name, err := base64.StdEncoding.DecodeString(p.Name)
	if err != nil {
		return err
	}
	p.Name = string(name)
	return nil
}

func Test_Sensitive(t *testing.T) {
	identityId := "adfadfa"
	so := &SensitiveObject{Name: "Alice", Age: 26}
	// 加密
	err := so.IntoSensitive(identityId)
	require.NoError(t, err)
	require.Equal(t, &SensitiveObject{Name: "QWxpY2U=", Age: 26}, so)

	// 解密
	err = so.FromSensitive(identityId)
	require.NoError(t, err)
	require.Equal(t, &SensitiveObject{Name: "Alice", Age: 26}, so)
}
