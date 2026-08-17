package token_limiter

import (
	"context"
	"time"
)

type AllowNRequest struct {
	Key   string    // 存储令牌的key
	Rate  int       // 令牌生成速率
	Burst int       // 桶的最大容量
	Now   time.Time // 当前时间, 零值表示使用 Redis 服务端时间
	N     int       // 请求的令牌数量
}

type TokenLimiterBackend interface {
	AllowN(ctx context.Context, r *AllowNRequest) (bool, error)
}
