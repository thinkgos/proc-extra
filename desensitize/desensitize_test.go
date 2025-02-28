package desensitize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type OriginalObject struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// 实现 DesensitizeAble[T any] 接口
func (p *OriginalObject) IntoDesensitized() *OriginalObject {
	// 脱敏操作，如对 Name 进行
	p.Name += " xxxx"
	return p
}

func Test_Desensitized(t *testing.T) {
	// 单个对象
	rawObj := &OriginalObject{Name: "Alice", Age: 26}
	// 切片
	rawObjSlices := []*OriginalObject{{Name: "Alice", Age: 26}, {Name: "Bob", Age: 27}}

	// 单个对
	deseObj := rawObj.IntoDesensitized()
	require.Equal(t, &OriginalObject{Name: "Alice xxxx", Age: 26}, deseObj)

	// 切片
	deseObjSlices := IntoDesensitized(rawObjSlices)
	require.Equal(t,
		[]*OriginalObject{{Name: "Alice xxxx", Age: 26}, {Name: "Bob xxxx", Age: 27}},
		deseObjSlices,
	)
}
