package verified

import (
	"context"
	"slices"
)

type TempGranter[S SceneValuer] interface {
	Issue(ctx context.Context, scene S, id string, opts ...Option) (string, error)
	Consume(ctx context.Context, scene S, id, token string) (bool, error)
}

// TempGrantGenerator the temp grant generator
type TempGrantGenerator interface {
	Name() string
	GenerateUniqueId() string
}

// TempGrant temp grant verifier
type TempGrant[S SceneValuer, P TempGrantGenerator, B StorageBackend] struct {
	p         P               // temp grant provider
	backend   B               // store backend
	keyPrefix string          // key prefix for captcha store
	param     *Param          // general param
	scenes    []SceneParam[S] // scene param.
}

// NewTempGrant new temp grant verifier instance.
func NewTempGrant[S SceneValuer, P TempGrantGenerator, B StorageBackend](p P, s B) *TempGrant[S, P, B] {
	return &TempGrant[S, P, B]{
		p:         p,
		backend:   s,
		keyPrefix: "temp-grant:ticket:",
		param:     NewParam(),
		scenes:    make([]SceneParam[S], 0),
	}
}

// Name the provider name
func (t *TempGrant[S, P, B]) Name() string { return t.p.Name() }

// SetKeyPrefix sets the key prefix.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (c *TempGrant[S, P, B]) SetKeyPrefix(keyPrefix string) *TempGrant[S, P, B] {
	c.keyPrefix = keyPrefix
	return c
}

// SetGeneralParam sets the general param.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (c *TempGrant[S, P, B]) SetGeneralParam(p *Param) *TempGrant[S, P, B] {
	c.param = p
	return c
}

// SetSceneParam sets the param for a specific scene.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (c *TempGrant[S, P, B]) SetSceneParam(scene S, param *Param) *TempGrant[S, P, B] {
	for i := range c.scenes {
		if c.scenes[i].scene == scene {
			c.scenes[i].param = param
			return c
		}
	}
	c.scenes = append(c.scenes, SceneParam[S]{scene: scene, param: param})
	c.scenes = slices.Clone(c.scenes)
	return c
}

func (c *TempGrant[S, P, B]) useScene(scene S, opts ...Option) *Param {
	p := c.param
	for i := range c.scenes {
		if c.scenes[i].scene == scene {
			p = c.scenes[i].param
			break
		}
	}
	return p.clone().apply(opts...)
}

// Issue a temp grant token. use option overwrite default param.
func (t *TempGrant[S, P, B]) Issue(ctx context.Context, scene S, id string, opts ...Option) (string, error) {
	p := t.useScene(scene, opts...)
	answer := t.p.GenerateUniqueId()
	err := t.backend.Save(ctx, &SaveArgs{
		Key:         t.formatKey(scene.Value(), id),
		KeyExpires:  p.KeyExpires,
		MaxAttempts: p.MaxAttempts,
		Answer:      answer,
	})
	if err != nil {
		return "", err
	}
	return answer, nil
}

// Consume the temp grant token.
func (t *TempGrant[S, P, B]) Consume(ctx context.Context, scene S, id, token string) (bool, error) {
	return t.backend.Verify(ctx, &VerifyArgs{
		Key:    t.formatKey(scene.Value(), id),
		Answer: token,
	})
}

func (c *TempGrant[S, P, B]) formatKey(scene, id string) string {
	return c.keyPrefix + scene + ":" + id
}
