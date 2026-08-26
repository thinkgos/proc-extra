package dual_token

import (
	"context"
	"time"
)

// DualToken 双令牌机制, 主要是 refresh token 管理和 access token的复用机制
type DualToken struct {
	backend   DualTokenBackend // 存储后端
	keyPrefix string           // 存储key前缀
	graceTTL  time.Duration    // access token 重用窗口时间
}

// NewDualToken 创建一个 DualToken 实例.
//
// 双令牌机制(access token + refresh token):
//   - access token: 用于授权访问, 无状态(JWT), 有效时间短(一般15min~60min), 过期后需通过 refresh token 换取新的.
//   - refresh token: 用于轮转 access token, 有状态(存储在后端, 如Redis), 有效时间长(一般7d~1m).
//
// 核心流程:
//   - Save: 用户登录成功后, 将 refresh token id 注册到后端白名单.
//   - Rotate: 客户端携带即将过期的 access token + refresh token 请求新凭证, 后端原子校验并轮转.
//   - Revoke: 用户登出时吊销凭证, 同时删除白名单和 grace 窗口.
//
// 默认配置:
//   - keyPrefix: "voucher:rt:", 可通过 SetKeyPrefix 自定义
//   - graceTTL: 10s, 可通过 SetGraceTTL 自定义. grace 窗口用于在轮转后短时间内允许旧 access token 继续使用, 避免并发请求失败.
//
// backend 参数为存储后端实现, 内置 Redis 实现见 NewRedisBackend.
func NewDualToken(backend DualTokenBackend) *DualToken {
	return &DualToken{
		backend:   backend,
		keyPrefix: "voucher:rt:",
		graceTTL:  time.Second * 10,
	}
}

func (d *DualToken) SetKeyPrefix(prefix string) {
	d.keyPrefix = prefix
}
func (d *DualToken) SetGraceTTL(ttl time.Duration) {
	d.graceTTL = ttl
}

type SaveParam struct {
	UserId         string    // M, user id
	SessionId      string    // M, session id
	RefreshTokenId string    // M, refresh token id
	Expires        time.Time // M, refresh token expires;
}

func (d *DualToken) Save(ctx context.Context, p *SaveParam) error {
	return d.backend.Save(ctx, &SaveRequest{
		Key:            d.formatKey(p.UserId, p.SessionId),
		RefreshTokenId: p.RefreshTokenId,
		TTL:            max(time.Until(p.Expires), time.Second),
	})
}

type RefreshParam struct {
	UserId                 string    // M, user id
	SessionId              string    // M, session id
	RequestRefreshTokenId  string    // M, old refresh token id
	NewRefreshTokenId      string    // M, new refresh token id
	NewRefreshToken        string    // M, new refresh token
	NewRefreshTokenExpires time.Time // M, new refresh token expires
	NewAccessToken         string    // M, new access token
	NewAccessTokenExpires  time.Time // M, new access token expires
}

type RotateResult struct {
	Success            bool
	AccessToken        string
	AccessTokenExpires time.Time
	RefreshToken       string
}

func (d *DualToken) Rotate(ctx context.Context, p *RefreshParam) (*RotateResult, error) {
	resp, err := d.backend.Rotate(ctx, &RotateRequest{
		Key:                   d.formatKey(p.UserId, p.SessionId),
		GraceKey:              d.formatResueKey(p.UserId, p.SessionId),
		RequestRefreshTokenId: p.RequestRefreshTokenId,
		NewRefreshTokenId:     p.NewRefreshTokenId,
		NewRefreshToken:       p.NewRefreshToken,
		NewRefreshTokenTTL:    max(time.Until(p.NewRefreshTokenExpires), time.Second),
		NewAccessToken:        p.NewAccessToken,
		NewAccessTokenExpires: p.NewAccessTokenExpires.Unix(),
		GraceTTL:              d.graceTTL,
	})
	if err != nil {
		return nil, err
	}
	return &RotateResult{
		Success:            resp.Success,
		AccessToken:        resp.AccessToken,
		AccessTokenExpires: time.Unix(resp.AccessTokenExpires, 0),
		RefreshToken:       resp.RefreshToken,
	}, nil
}

type RevokeParam struct {
	UserId    string // M, user id
	SessionId string // M, session id
}

func (d *DualToken) Revoke(ctx context.Context, p *RevokeParam) error {
	return d.backend.Revoke(ctx, &RevokeRequest{
		Key:      d.formatKey(p.UserId, p.SessionId),
		GraceKey: d.formatResueKey(p.UserId, p.SessionId),
	})
}

func (d *DualToken) formatKey(userId, sessionId string) string {
	return d.keyPrefix + userId + ":" + sessionId
}

func (d *DualToken) formatResueKey(userId, sessionId string) string {
	return d.keyPrefix + userId + ":" + sessionId + ":grace"
}
