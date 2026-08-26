# 双令牌方案 (Dual Token)

## 概述

双令牌方案是一种安全的认证刷新机制, 使用两个令牌协同工作:

- **access token**: 短有效期(15min~60min), 无状态(JWT), 用于授权 API 访问.
- **refresh token**: 长有效期(7d~1m), 有状态(存储在后端), 用于在 access token 过期后换取新的凭证对.

相比单 token 方案, 双令牌方案的优势:
1. access token 有效期短, 即使泄露, 风险窗口有限.
2. refresh token 存储在后端, 可随时吊销(如用户登出、密码修改).
3. 轮转(refresh token rotation)机制防止 refresh token 被重复使用.

## 存储结构

使用两个 Redis key 管理凭证状态:

```
key1: voucher:rt:{userId}:{sessionId}       -> refresh token id (白名单)
key2: voucher:rt:{userId}:{sessionId}:grace  -> {rt_id: xx, rt: xx, at: xx, at_exp: xx} (grace 窗口)
```

- **key1 (白名单)**: 存储当前有效的 refresh token id. 轮转时校验客户端提交的 refresh token id 是否与之匹配.
- **key2 (grace 窗口)**: 轮转成功后, 旧的 access token 和 refresh token 会在 grace 窗口内短暂保留, 供并发请求复用, 避免因竞态条件导致请求失败.
- **userId**: 用户id
- **sessionId**: 会话id, 也就是当前用户登陆后, 这一对双令牌的会话id, 后续轮转均以此会话id作为绑定关系.

## 核心流程

### 1. 登录保存 (Save)

用户登录成功后, 调用 `Save` 将 refresh token id 注册到白名单:

```
Save(ctx, &SaveParam{
    UserId:         "user1",        // 用户id
    SessionId:      "sess1",        // 会话id
    RefreshTokenId: "rt-xxx",       // refresh token 的 jti
    Expires:        7天后过期,       // refresh token 的过期时间
})
```

底层操作: `SET voucher:rt:user1:sess1 "rt-xxx" EX 604800`

### 2. 轮转凭证 (Rotate)

客户端的 access token 即将过期时, 携带 refresh token 请求新凭证:

```
Rotate(ctx, &RefreshParam{
    UserId:                 "user1",        // 用户id
    SessionId:              "sess1",        // 会话id
    RequestRefreshTokenId:  "old-rt-jti",   // 当前 refresh token 的 jti
    NewRefreshTokenId:      "new-rt-jti",   // 新 refresh token 的 jti
    NewRefreshToken:        "new-rt-token", // 新 refresh token
    NewRefreshTokenExpires: 7天后过期,       // 新 refresh token 的过期时间
    NewAccessToken:         "new-at-token", // 新 access token
    NewAccessTokenExpires:  15分钟后过期,     // 新 access token 的过期时间
})
```

底层通过 Lua 脚本原子执行:

1. 校验 `RequestRefreshTokenId` 是否与白名单中存储的一致.
2. 若匹配:
   - 更新白名单为 `NewRefreshTokenId`.
   - 将旧的 access token / refresh token 写入 grace 窗口(短期有效).
   - 返回新的凭证对.
3. 若不匹配(可能已被并发请求轮转):
   - 从 grace 窗口读取凭证返回(复用窗口).
   - `Success` 返回 `false`, 调用方可根据业务决定是否需要重新登录.

### 3. 吊销凭证 (Revoke)

用户登出或需要强制下线时, 调用 `Revoke` 删除白名单和 grace 窗口:

```
Revoke(ctx, &RevokeParam{
    UserId:    "user1",
    SessionId: "sess1",
})
```

底层操作: `DEL voucher:rt:user1:sess1 voucher:rt:user1:sess1:grace`

## 接口说明

### DualToken (业务层)

| 方法 | 说明 |
|------|------|
| `NewDualToken(backend)` | 创建实例, 需传入存储后端 |
| `SetKeyPrefix(prefix)` | 自定义存储 key 前缀, 默认 `"voucher:rt:"` |
| `SetGraceTTL(ttl)` | 自定义 grace 窗口时长, 默认 `10s` |
| `Save(ctx, param)` | 登录后保存 refresh token 白名单 |
| `Rotate(ctx, param)` | 轮转凭证(access token + refresh token) |
| `Revoke(ctx, param)` | 吊销凭证(登出/强制下线) |

### DualTokenBackend (存储层)

| 方法 | 说明 |
|------|------|
| `Save(ctx, req)` | 存储 refresh token id |
| `Rotate(ctx, req)` | 原子轮转凭证(校验 + 替换 + grace) |
| `Revoke(ctx, req)` | 删除白名单和 grace 窗口 |

内置实现: `NewRedisBackend(client)` — 基于 Redis + Lua 脚本, 保证原子性.

## 使用示例

```go
import (
    "github.com/redis/go-redis/v9"
    "github.com/thinkgos/admin-go/pkg/core/dual_token"
)

// 初始化
client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
dt := dual_token.NewDualToken(dual_token.NewRedisBackend(client))

// 可选: 自定义配置
dt.SetKeyPrefix("myapp:auth:")
dt.SetGraceTTL(15 * time.Second)

// 登录后保存
dt.Save(ctx, &dual_token.SaveParam{
    UserId:         userId,
    SessionId:      sessionId,
    RefreshTokenId: refreshTokenJTI,
    Expires:        time.Now().Add(7 * 24 * time.Hour),
})

// 轮转凭证
result, err := dt.Rotate(ctx, &dual_token.RefreshParam{
    UserId:                 userId,
    SessionId:              sessionId,
    RequestRefreshTokenId:  oldRefreshTokenJTI,
    NewRefreshTokenId:      newRefreshTokenJTI,
    NewRefreshToken:        newRefreshToken,
    NewRefreshTokenExpires: time.Now().Add(7 * 24 * time.Hour),
    NewAccessToken:         newAccessToken,
    NewAccessTokenExpires:  time.Now().Add(15 * time.Minute),
})

// 登出吊销
dt.Revoke(ctx, &dual_token.RevokeParam{
    UserId:    userId,
    SessionId: sessionId,
})
```

## 并发安全

轮转操作通过 Redis Lua 脚本保证原子性. 当多个并发请求同时携带相同的 refresh token 发起轮转时:

1. 只有第一个请求会成功轮转(`Success=true`), 获得全新的凭证对.
2. 后续请求在 grace 窗口内会收到上一次轮转的凭证(`Success=true`), 调用方可直接使用.
3. grace 窗口过期后, 使用旧的`refresh token`请求将失败, 客户端需要重新登录.
