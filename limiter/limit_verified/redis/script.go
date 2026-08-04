package redis

import _ "embed"

const (
	InnerLimitVerifiedEvaluate_Success       = 0
	InnerLimitVerifiedEvaluate_OverQuota     = 1
	InnerLimitVerifiedEvaluate_TooFrequently = 2
)

const (
	InnerLimitVerifiedVerify_Success = 0
	InnerLimitVerifiedVerify_Failure = 1
	InnerLimitVerifiedVerify_Expired = 2
)

var (
	//go:embed limit_verified_evaluate.lua
	ScriptLimitVerifiedEvaluate string
	//go:embed limit_verified_rollback.lua
	ScriptLimitVerifiedRollback string
	//go:embed limit_verified_verify_code.lua
	LimitVerifiedVerifyCodeScript string
)
