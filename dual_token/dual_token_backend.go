package dual_token

import (
	"context"
	"time"
)

// SaveRequest 保存 refresh token 白名单请求
type SaveRequest struct {
	Key            string        // 存储key, 格式: {keyPrefix}{userId}:{sessionId}
	RefreshTokenId string        // refresh token id (jti), 作为白名单凭证
	TTL            time.Duration // refresh token 过期时间, 到期后自动从存储中移除
}

// RotateRequest 轮转凭证请求
type RotateRequest struct {
	Key                   string        // 存储key, 格式: {keyPrefix}{userId}:{sessionId}
	GraceKey              string        // grace 窗口key, 格式: {keyPrefix}{userId}:{sessionId}:grace
	RequestRefreshTokenId string        // 客户端提交的旧 refresh token id, 用于校验是否为当前有效凭证
	NewRefreshTokenId     string        // 新 refresh token id (jti), 用于替换旧白名单
	NewRefreshToken       string        // 新 refresh token 字符串, 用于存入 grace 窗口供并发请求复用
	NewRefreshTokenTTL    time.Duration // 新 refresh token 过期时间
	NewAccessToken        string        // 新 access token 字符串, 用于存入 grace 窗口
	NewAccessTokenExpires int64         // 新 access token 过期时间(unix timestamp)
	GraceTTL              time.Duration // grace 窗口时长, 轮转成功后旧 access token 在此窗口内仍可复用
}

// RotateReply 轮转凭证响应
type RotateReply struct {
	Success            bool   // 轮转是否成功, false, 表示 refresh token id 不匹配
	AccessToken        string // 当前有效的 access token(若在 grace 窗口内, 可能返回旧的)
	AccessTokenExpires int64  // access token 过期时间(unix timestamp)
	RefreshToken       string // 当前有效的 refresh token(若在 grace 窗口内, 可能返回旧的)
}

// RevokeRequest 吊销凭证请求
type RevokeRequest struct {
	Key      string // 存储key, 格式: {keyPrefix}{userId}:{sessionId}
	GraceKey string // grace 窗口key, 格式: {keyPrefix}{userId}:{sessionId}:grace
}

// DualTokenBackend 双令牌存储后端接口.
//
// 内置实现: NewRedisBackend (基于 Redis + Lua 脚本, 保证原子性).
// 可自行实现此接口接入其他存储(如 MySQL、etcd 等).
type DualTokenBackend interface {
	// Save 保存 refresh token 白名单, 即注册 refresh token id 到存储中, 后续轮转时校验使用.
	Save(ctx context.Context, r *SaveRequest) error
	// Rotate 原子性地轮转凭证: 校验旧 refresh token id -> 替换为新 refresh token -> 存入 grace 窗口.
	// 若 refresh token id 不匹配则返回 Success=false.
	Rotate(ctx context.Context, r *RotateRequest) (*RotateReply, error)
	// Revoke 吊销凭证, 删除 refresh token 白名单和 grace 窗口, 使该会话立即失效.
	Revoke(ctx context.Context, r *RevokeRequest) error
}
