package verified

import (
	"context"
	"errors"
	"slices"
)

type CaptchaVerifier[S SceneValuer] interface {
	Generate(ctx context.Context, driverName string, scene S, opts ...Option) (id, question string, err error)
	Verify(ctx context.Context, scene S, id, answer string) (bool, error)
}

// Challenge question and answer.
type Challenge struct {
	Id       string
	Question string
	Answer   string
}

// ChallengeProvider the captcha challenge provider
type ChallengeProvider interface {
	Name() string
	GenerateChallenge(ctx context.Context) (*Challenge, error)
}

// CaptchaDriver the captcha driver
type CaptchaDriver interface {
	Driver(dName string) ChallengeProvider
}

// Captcha verified captcha limit
type Captcha[S SceneValuer, P CaptchaDriver, B StorageBackend] struct {
	p         P               // captcha provider
	backend   B               // store backend
	keyPrefix string          // key prefix for captcha store
	param     *Param          // general param
	scenes    []SceneParam[S] // scene param.
}

// NewCaptcha new captcha instance.
func NewCaptcha[S SceneValuer, P CaptchaDriver, B StorageBackend](p P, backend B) *Captcha[S, P, B] {
	return &Captcha[S, P, B]{
		p:         p,
		backend:   backend,
		keyPrefix: "captcha:scene:",
		param:     NewParam(),
		scenes:    make([]SceneParam[S], 0),
	}
}

// SetKeyPrefix sets the key prefix for the limit verified.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (c *Captcha[S, P, B]) SetKeyPrefix(keyPrefix string) *Captcha[S, P, B] {
	c.keyPrefix = keyPrefix
	return c
}

// SetParam sets the general param for the limit verified.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (c *Captcha[S, P, B]) SetParam(p *Param) *Captcha[S, P, B] {
	c.param = p
	return c
}

// SetSceneParam sets the param for a specific scene.
// NOTE: This method is NOT safe for concurrent use. It should only be called during initialization.
func (c *Captcha[S, P, B]) SetSceneParam(scene S, param *Param) *Captcha[S, P, B] {
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

func (c *Captcha[S, P, B]) useScene(scene S, opts ...Option) *Param {
	p := c.param
	for i := range c.scenes {
		if c.scenes[i].scene == scene {
			p = c.scenes[i].param
			break
		}
	}
	return p.clone().apply(opts...)
}

// Name the provider name
func (c *Captcha[S, P, B]) Name(driverName string) string {
	return c.p.Driver(driverName).Name()
}

// Generate generate id, question.
func (c *Captcha[S, P, B]) Generate(ctx context.Context, driverName string, scene S, opts ...Option) (id, question string, err error) {
	qa, err := c.p.Driver(driverName).GenerateChallenge(ctx)
	if err != nil {
		return "", "", err
	}
	p := c.useScene(scene)
	err = c.backend.Save(ctx, &SaveArgs{
		Key:         c.formatKey(scene.Value(), qa.Id),
		KeyExpires:  p.KeyExpires,
		MaxAttempts: p.MaxAttempts,
		Answer:      qa.Answer,
	})
	if err != nil {
		return "", "", err
	}
	return qa.Id, qa.Question, nil
}

// Verify the answer.
func (c *Captcha[S, P, B]) Verify(ctx context.Context, scene S, id, answer string) (bool, error) {
	return c.backend.Verify(ctx, &VerifyArgs{
		Key:    c.formatKey(scene.Value(), id),
		Answer: answer,
	})
}

func (c *Captcha[S, P, B]) formatKey(scene, id string) string {
	return c.keyPrefix + scene + ":" + id
}

type UnsupportedChallengeProvider struct{}

func (x UnsupportedChallengeProvider) Name() string {
	return "Unsupported captcha driver"
}
func (x UnsupportedChallengeProvider) GenerateChallenge(ctx context.Context) (*Challenge, error) {
	return nil, errors.New(x.Name())
}
