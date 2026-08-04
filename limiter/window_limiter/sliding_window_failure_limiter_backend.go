package window_limiter

import "context"

type FailureLimiterEvaluateRequest struct {
	Key         string // key
	Window      int    // sliding window size in seconds
	MaxFailures int    // max failures in the sliding window
	UniqueId    string // unique id.
	IsFailure   bool   // whether the attempt is failure?
}
type FailureLimiterCheckRequest struct {
	Key         string // key
	Window      int    // sliding window size in seconds
	MaxFailures int    // max failures in the sliding window
}
type FailureLimiterLockRequest struct {
	Key         string // key
	Window      int    // sliding window size in seconds
	MaxFailures int    // max failures in the sliding window
}
type FailureLimiterResult struct {
	// whether the operation is allowed or not.
	// Evaluate: whether current operation is allowed or not.
	// Check: whether next operation is allowed or not.
	Allow       bool
	ExpireAt    int64 // unix timestamp (seconds) at which the current window fully resets.
	Failures    int   // the current count of failures in the sliding window
	MaxFailures int   // the max failures in the sliding window
}

// SlidingWindowFailureLimiterBackend 滑动窗口失败限制器后端.
type SlidingWindowFailureLimiterBackend interface {
	Evaluate(ctx context.Context, v *FailureLimiterEvaluateRequest) (*FailureLimiterResult, error)
	Check(ctx context.Context, v *FailureLimiterCheckRequest) (*FailureLimiterResult, error)
	Lock(ctx context.Context, v *FailureLimiterLockRequest) (*FailureLimiterResult, error)
	Reset(ctx context.Context, key string) error
}
