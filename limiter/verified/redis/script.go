package redis

import _ "embed"

//go:embed save.lua
var ScriptSave string

//go:embed verify.lua
var ScriptVerify string
