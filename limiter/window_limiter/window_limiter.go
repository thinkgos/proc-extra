package window_limiter

type SceneValuer interface {
	comparable
	Value() string
}
