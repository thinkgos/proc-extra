package redis

import (
	_ "embed"
)

var (
	//go:embed sliding_window_limiter_take.lua
	ScriptSlidingWindowLimiterTake string
	//go:embed sliding_window_limiter_lock.lua
	ScriptSlidingWindowLimiterLock string
	//go:embed sliding_window_limiter_check.lua
	ScriptSlidingWindowLimiterCheck string
)

var (
	//go:embed sliding_window_failure_limiter_evaluate.lua
	ScriptSlidingWindowFailureLimiterEvaluate string
	//go:embed sliding_window_failure_limiter_lock.lua
	ScriptSlidingWindowFailureLimiterLock string
	//go:embed sliding_window_failure_limiter_check.lua
	ScriptSlidingWindowFailureLimiterCheck string
)
