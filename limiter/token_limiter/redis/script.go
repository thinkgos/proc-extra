package redis

import (
	_ "embed"
)

//go:embed token_rate.lua
var ScriptTokenRate string
