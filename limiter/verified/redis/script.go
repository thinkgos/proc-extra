package redis

import _ "embed"

var (
	//go:embed save.lua
	ScriptSave string
	//go:embed verify.lua
	ScriptVerify string
	//go:embed inspect.lua
	ScriptInspect string
)
