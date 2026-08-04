package window_limiter

import "context"

type LimiterTakeRequest struct {
	Key      string // key
	Window   int    // sliding window size in seconds
	MaxLimit int    // max limit requests in the sliding window
	UniqueId string // unique id.
}
type LimiterCheckRequest struct {
	Key      string // key
	Window   int    // sliding window size in seconds
	MaxLimit int    // max limit requests in the sliding window
}
type LimiterLockRequest struct {
	Key      string // key
	Window   int    // sliding window size in seconds
	MaxLimit int    // max limit requests in the sliding window
}

type LimiterResult struct {
	// whether the request is allowed or not.
	// Take: whether current request is allowed or not.
	// Check: whether next request is allowed or not.
	Allow    bool
	ExpireAt int64 // unix timestamp (seconds) at which the current window fully resets.
	Count    int   // the current count of requests in the sliding window
	MaxLimit int   // the max limit requests in the sliding window
}
type SlidingWindowLimiterBackend interface {
	Take(ctx context.Context, v *LimiterTakeRequest) (*LimiterResult, error)
	Check(ctx context.Context, v *LimiterCheckRequest) (*LimiterResult, error)
	Lock(ctx context.Context, v *LimiterLockRequest) (*LimiterResult, error)
	Reset(ctx context.Context, key string) error
}
