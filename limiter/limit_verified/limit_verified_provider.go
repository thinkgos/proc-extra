package limit_verified

import (
	"context"
)

// LimitVerifiedProvider the provider
type LimitVerifiedProvider interface {
	Name() string
	SendCode(ctx context.Context, target, code string) error
}

var _ LimitVerifiedProvider = DummyDriver{}

type DummyDriver struct{}

func (DummyDriver) Name() string                                            { return "limit-verified-dummy-provider" }
func (DummyDriver) SendCode(ctx context.Context, target, code string) error { return nil }
