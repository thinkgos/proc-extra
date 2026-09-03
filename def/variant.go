package def

type VariantAble[T any] interface {
	IntoVariant() T
}

func IntoVariant[T VariantAble[R], R any](s []T) []R {
	r := make([]R, 0, len(s))
	for _, v := range s {
		r = append(r, v.IntoVariant())
	}
	return r
}
