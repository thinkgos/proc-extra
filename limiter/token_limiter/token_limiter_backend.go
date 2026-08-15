package token_limiter

import (
	"context"
	"time"
)

type AllowNRequest struct {
	Key   string
	Rate  int
	Burst int
	Now   time.Time
	N     int
}

type TokenRateBackend interface {
	AllowN(ctx context.Context, r *AllowNRequest) (bool, error)
}
