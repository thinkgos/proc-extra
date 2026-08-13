package redis

import _ "embed"

const (
	InnerLimitVerifiedEvaluate_Success       = 0 // 成功
	InnerLimitVerifiedEvaluate_OverQuota     = 1 // 超过配额
	InnerLimitVerifiedEvaluate_TooFrequently = 2 // 过于频繁
)

const (
	InnerLimitVerifiedVerify_Success = 0 // 成功
	InnerLimitVerifiedVerify_Failure = 1 // 失败
	InnerLimitVerifiedVerify_Expired = 2 // 已失效
)

var (
	//go:embed limit_verified_evaluate.lua
	ScriptLimitVerifiedEvaluate string
	//go:embed limit_verified_rollback.lua
	ScriptLimitVerifiedRollback string
	//go:embed limit_verified_verify_code.lua
	LimitVerifiedVerifyCodeScript string
)
