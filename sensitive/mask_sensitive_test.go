package sensitive

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type TestOriginalObject struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// 实现 MaskSensitiveAble[T any] 接口
func (p *TestOriginalObject) MaskSensitive() *TestOriginalObject {
	// 脱敏操作，如对 Name 进行
	p.Name += " xxxx"
	return p
}

func Test_MaskSensitive(t *testing.T) {
	// 单个对象
	rawObj := &TestOriginalObject{Name: "Alice", Age: 26}
	// 切片
	rawObjSlices := []*TestOriginalObject{{Name: "Alice", Age: 26}, {Name: "Bob", Age: 27}}

	// 单个对象脱敏
	maskObj := rawObj.MaskSensitive()
	require.Equal(t, &TestOriginalObject{Name: "Alice xxxx", Age: 26}, maskObj)

	// 切片整体脱敏
	deseObjSlices := MaskSensitive(rawObjSlices)
	require.Equal(t,
		[]*TestOriginalObject{{Name: "Alice xxxx", Age: 26}, {Name: "Bob xxxx", Age: 27}},
		deseObjSlices,
	)
}
