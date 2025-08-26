package sensitive

const SixStar = "******"

// 定义脱敏接口
type MaskSensitiveAble[T any] interface {
	MaskSensitive() T
}

// MaskSensitive 切片数据, 统一脱敏
func MaskSensitive[T MaskSensitiveAble[R], R any](collection []T) []R {
	result := make([]R, len(collection))
	for i := range collection {
		result[i] = collection[i].MaskSensitive()
	}
	return result
}
