package window_limiter

import "slices"

// SlidingWindowLimiterParam sliding window limiter param.
type SlidingWindowLimiterParam struct {
	Window   int // sliding window in seconds
	MaxLimit int // max requests/failures in the sliding window
}

type SceneParam[S SceneValuer, P any] struct {
	scene S
	param *P
}

type SceneParamRegistry[S SceneValuer, P any] struct {
	keyPrefix string
	param     *P
	scenes    []SceneParam[S, P]
}

// NewSceneParamRegistry new a SceneParamRegistry instance.
func NewSceneParamRegistry[S SceneValuer, P any](keyPrefix string, param *P) SceneParamRegistry[S, P] {
	return SceneParamRegistry[S, P]{
		keyPrefix: keyPrefix,
		param:     param,
		scenes:    make([]SceneParam[S, P], 0),
	}
}

// SetKeyPrefix sets the key prefix.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (l *SceneParamRegistry[S, P]) SetKeyPrefix(keyPrefix string) {
	l.keyPrefix = keyPrefix
}

// SetGeneralParam sets the default param.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (l *SceneParamRegistry[S, P]) SetGeneralParam(p *P) {
	l.param = p
}

// SetSceneParam sets the param for a specific scene.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (l *SceneParamRegistry[S, P]) SetSceneParam(scene S, param *P) {
	for i := range l.scenes {
		if l.scenes[i].scene == scene {
			l.scenes[i].param = param
			return
		}
	}
	l.scenes = append(l.scenes, SceneParam[S, P]{scene: scene, param: param})
	l.scenes = slices.Clone(l.scenes)
}

func (l *SceneParamRegistry[S, P]) useScene(scene S) *P {
	for i := range l.scenes {
		if l.scenes[i].scene == scene {
			return l.scenes[i].param
		}
	}
	return l.param
}

func (l *SceneParamRegistry[S, P]) formatKey(scene, id string) string {
	return l.keyPrefix + scene + ":" + id
}

func (l *SceneParamRegistry[S, P]) formatLockedKey(scene, id string) string {
	return l.keyPrefix + scene + ":" + id + ":_locked"
}
