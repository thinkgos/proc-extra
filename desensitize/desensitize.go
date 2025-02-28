package desensitize

// 定义脱敏接口
type DesensitizeAble[T any] interface {
	IntoDesensitized() T
}

// 切片数据, 统一脱敏
func IntoDesensitized[T DesensitizeAble[R], R any](collection []T) []R {
	result := make([]R, len(collection))
	for i := range collection {
		result[i] = collection[i].IntoDesensitized()
	}
	return result
}
