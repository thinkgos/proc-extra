package verified

import (
	"context"
	"time"
)

// SaveArgs store arguments
type SaveArgs struct {
	Key         string
	KeyExpires  time.Duration
	MaxAttempts int
	Answer      string
}

// VerifyArgs verify arguments
type VerifyArgs struct {
	Key    string
	Answer string
}

// StorageBackend store engine
type StorageBackend interface {
	Save(context.Context, *SaveArgs) error
	Verify(context.Context, *VerifyArgs) (bool, error)
}
