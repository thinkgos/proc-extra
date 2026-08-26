package v9

import (
	"context"
	"errors"
	"time"

	_ "embed"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
	"github.com/thinkgos/proc-extra/dual_token"
)

//go:embed rotate.lua
var scriptRotate string

// RedisStore implements DualTokenBackend using Redis.
type RedisStore struct {
	client *redis.Client
}

// NewRedisBackend creates a new RedisBackend instance.
func NewRedisBackend(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
}

// Store 保存 refresh token 白名单
func (b *RedisStore) Save(ctx context.Context, r *dual_token.SaveRequest) error {
	return b.client.Set(ctx, r.Key, r.RefreshTokenId, r.TTL).Err()
}

// Rotate 真正刷新凭证, old access token id, old refresh token id 同时满足才能真正刷新成功
func (b *RedisStore) Rotate(ctx context.Context, r *dual_token.RotateRequest) (*dual_token.RotateReply, error) {
	vals, err := b.client.Eval(
		ctx,
		scriptRotate,
		[]string{
			r.Key,
			r.GraceKey,
		},
		r.RequestRefreshTokenId,
		r.NewRefreshTokenId,
		r.NewRefreshToken,
		r.NewAccessToken,
		r.NewAccessTokenExpires,
		r.NewRefreshTokenTTL/time.Second,
		r.GraceTTL/time.Second,
	).Slice()
	if err != nil {
		return nil, err
	}
	if len(vals) != 4 {
		return nil, errors.New("lua script rotate invalid result")
	}
	return &dual_token.RotateReply{
		Success:            cast.ToInt64(vals[0]) == 0,
		AccessToken:        cast.ToString(vals[1]),
		AccessTokenExpires: cast.ToInt64(vals[2]),
		RefreshToken:       cast.ToString(vals[3]),
	}, nil
}

func (b *RedisStore) Revoke(ctx context.Context, r *dual_token.RevokeRequest) error {
	return b.client.Del(ctx, r.Key, r.GraceKey).Err()
}
