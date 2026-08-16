package redis

import (
	_ "embed"
)

//go:embed token_limiter.lua
var ScriptTokenLimiter string
