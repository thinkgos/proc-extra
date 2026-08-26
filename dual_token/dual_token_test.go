package dual_token

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockBackend implements DualTokenBackend for testing
type mockBackend struct {
	saveFn   func(ctx context.Context, r *SaveRequest) error
	rotateFn func(ctx context.Context, r *RotateRequest) (*RotateReply, error)
	revokeFn func(ctx context.Context, r *RevokeRequest) error

	saveCalls   []*SaveRequest
	rotateCalls []*RotateRequest
	revokeCalls []*RevokeRequest
}

func (m *mockBackend) Save(ctx context.Context, r *SaveRequest) error {
	m.saveCalls = append(m.saveCalls, r)
	if m.saveFn != nil {
		return m.saveFn(ctx, r)
	}
	return nil
}

func (m *mockBackend) Rotate(ctx context.Context, r *RotateRequest) (*RotateReply, error) {
	m.rotateCalls = append(m.rotateCalls, r)
	if m.rotateFn != nil {
		return m.rotateFn(ctx, r)
	}
	return nil, nil
}

func (m *mockBackend) Revoke(ctx context.Context, r *RevokeRequest) error {
	m.revokeCalls = append(m.revokeCalls, r)
	if m.revokeFn != nil {
		return m.revokeFn(ctx, r)
	}
	return nil
}

// NewDualToken 构造函数测试
func TestNewDualToken(t *testing.T) {
	backend := &mockBackend{}
	dt := NewDualToken(backend)

	if dt.backend != backend {
		t.Fatal("backend 未正确赋值")
	}
	if dt.keyPrefix != "voucher:rt:" {
		t.Fatalf("keyPrefix 期望 'voucher:rt:', 实际 '%s'", dt.keyPrefix)
	}
	if dt.graceTTL != 10*time.Second {
		t.Fatalf("graceTTL 期望 10s, 实际 %s", dt.graceTTL)
	}
}

// SetKeyPrefix 测试
func TestSetKeyPrefix(t *testing.T) {
	backend := &mockBackend{}
	dt := NewDualToken(backend)

	dt.SetKeyPrefix("custom:prefix:")
	if dt.keyPrefix != "custom:prefix:" {
		t.Fatalf("keyPrefix 期望 'custom:prefix:', 实际 '%s'", dt.keyPrefix)
	}
}

// SetGraceTTL 测试
func TestSetGraceTTL(t *testing.T) {
	backend := &mockBackend{}
	dt := NewDualToken(backend)

	dt.SetGraceTTL(30 * time.Second)
	if dt.graceTTL != 30*time.Second {
		t.Fatalf("graceTTL 期望 30s, 实际 %s", dt.graceTTL)
	}
}

// Save 测试 - 正常保存
func TestSave(t *testing.T) {
	backend := &mockBackend{}
	dt := NewDualToken(backend)

	expires := time.Now().Add(7 * 24 * time.Hour)
	p := &SaveParam{
		UserId:         "user1",
		SessionId:      "sess1",
		RefreshTokenId: "rtid1",
		Expires:        expires,
	}

	err := dt.Save(context.Background(), p)
	if err != nil {
		t.Fatalf("Save 不应返回错误: %v", err)
	}

	if len(backend.saveCalls) != 1 {
		t.Fatalf("期望调用 Save 1次, 实际 %d次", len(backend.saveCalls))
	}

	req := backend.saveCalls[0]
	if req.Key != "voucher:rt:user1:sess1" {
		t.Fatalf("Key 期望 'voucher:rt:user1:sess1', 实际 '%s'", req.Key)
	}
	if req.RefreshTokenId != "rtid1" {
		t.Fatalf("RefreshTokenId 期望 'rtid1', 实际 '%s'", req.RefreshTokenId)
	}
	if req.TTL < time.Second {
		t.Fatal("TTL 不应小于 1s")
	}
}

// Save 测试 - 后端返回错误
func TestSave_Error(t *testing.T) {
	wantErr := errors.New("redis error")
	backend := &mockBackend{
		saveFn: func(ctx context.Context, r *SaveRequest) error {
			return wantErr
		},
	}
	dt := NewDualToken(backend)

	err := dt.Save(context.Background(), &SaveParam{
		UserId:         "user1",
		SessionId:      "sess1",
		RefreshTokenId: "rtid1",
		Expires:        time.Now().Add(time.Hour),
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("期望错误 %v, 实际 %v", wantErr, err)
	}
}

// Save 测试 - 过期时间已过的场景, TTL应至少为1s
func TestSave_ExpiredTTL(t *testing.T) {
	backend := &mockBackend{}
	dt := NewDualToken(backend)

	err := dt.Save(context.Background(), &SaveParam{
		UserId:         "user1",
		SessionId:      "sess1",
		RefreshTokenId: "rtid1",
		Expires:        time.Now().Add(-time.Hour), // 已过期
	})
	if err != nil {
		t.Fatalf("Save 不应返回错误: %v", err)
	}

	req := backend.saveCalls[0]
	if req.TTL != time.Second {
		t.Fatalf("TTL 期望 1s (最小值), 实际 %s", req.TTL)
	}
}

// Rotate 测试 - 成功轮转
func TestRotate_Success(t *testing.T) {
	now := time.Now()
	backend := &mockBackend{
		rotateFn: func(ctx context.Context, r *RotateRequest) (*RotateReply, error) {
			return &RotateReply{
				Success:            true,
				AccessToken:        "new-at",
				AccessTokenExpires: now.Add(15 * time.Minute).Unix(),
				RefreshToken:       "new-rt",
			}, nil
		},
	}
	dt := NewDualToken(backend)

	resp, err := dt.Rotate(context.Background(), &RefreshParam{
		UserId:                 "user1",
		SessionId:              "sess1",
		RequestRefreshTokenId:  "old-rtid",
		NewRefreshTokenId:      "new-rtid",
		NewRefreshToken:        "new-rt",
		NewRefreshTokenExpires: now.Add(7 * 24 * time.Hour),
		NewAccessToken:         "new-at",
		NewAccessTokenExpires:  now.Add(15 * time.Minute),
	})
	if err != nil {
		t.Fatalf("Rotate 不应返回错误: %v", err)
	}
	if !resp.Success {
		t.Fatal("期望 Success=true")
	}
	if resp.AccessToken != "new-at" {
		t.Fatalf("AccessToken 期望 'new-at', 实际 '%s'", resp.AccessToken)
	}
	if resp.RefreshToken != "new-rt" {
		t.Fatalf("RefreshToken 期望 'new-rt', 实际 '%s'", resp.RefreshToken)
	}

	// 验证请求参数
	if len(backend.rotateCalls) != 1 {
		t.Fatalf("期望调用 Rotate 1次, 实际 %d次", len(backend.rotateCalls))
	}
	req := backend.rotateCalls[0]
	if req.Key != "voucher:rt:user1:sess1" {
		t.Fatalf("Key 期望 'voucher:rt:user1:sess1', 实际 '%s'", req.Key)
	}
	if req.GraceKey != "voucher:rt:user1:sess1:grace" {
		t.Fatalf("GraceKey 期望 'voucher:rt:user1:sess1:grace', 实际 '%s'", req.GraceKey)
	}
	if req.NewRefreshTokenId != "new-rtid" {
		t.Fatalf("NewRefreshTokenId 期望 'new-rtid', 实际 '%s'", req.NewRefreshTokenId)
	}
	if req.NewRefreshToken != "new-rt" {
		t.Fatalf("NewRefreshToken 期望 'new-rt', 实际 '%s'", req.NewRefreshToken)
	}
	if req.NewAccessToken != "new-at" {
		t.Fatalf("NewAccessToken 期望 'new-at', 实际 '%s'", req.NewAccessToken)
	}
	if req.GraceTTL != 10*time.Second {
		t.Fatalf("GraceTTL 期望 10s, 实际 %s", req.GraceTTL)
	}
	if req.NewRefreshTokenTTL < time.Second {
		t.Fatal("NewRefreshTokenTTL 不应小于 1s")
	}
}

// Rotate 测试 - 失败轮转 (Success=false)
func TestRotate_Failure(t *testing.T) {
	backend := &mockBackend{
		rotateFn: func(ctx context.Context, r *RotateRequest) (*RotateReply, error) {
			return &RotateReply{
				Success: false,
			}, nil
		},
	}
	dt := NewDualToken(backend)

	resp, err := dt.Rotate(context.Background(), &RefreshParam{
		UserId:                 "user1",
		SessionId:              "sess1",
		RequestRefreshTokenId:  "old-rtid",
		NewRefreshTokenId:      "new-rtid",
		NewRefreshToken:        "new-rt",
		NewRefreshTokenExpires: time.Now().Add(time.Hour),
		NewAccessToken:         "new-at",
		NewAccessTokenExpires:  time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		t.Fatalf("Rotate 不应返回错误: %v", err)
	}
	if resp.Success {
		t.Fatal("期望 Success=false")
	}
}

// Rotate 测试 - 后端返回错误
func TestRotate_Error(t *testing.T) {
	wantErr := errors.New("rotate failed")
	backend := &mockBackend{
		rotateFn: func(ctx context.Context, r *RotateRequest) (*RotateReply, error) {
			return nil, wantErr
		},
	}
	dt := NewDualToken(backend)

	resp, err := dt.Rotate(context.Background(), &RefreshParam{
		UserId:                 "user1",
		SessionId:              "sess1",
		RequestRefreshTokenId:  "old-rtid",
		NewRefreshTokenId:      "new-rtid",
		NewRefreshToken:        "new-rt",
		NewRefreshTokenExpires: time.Now().Add(time.Hour),
		NewAccessToken:         "new-at",
		NewAccessTokenExpires:  time.Now().Add(15 * time.Minute),
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("期望错误 %v, 实际 %v", wantErr, err)
	}
	if resp != nil {
		t.Fatal("错误时 resp 应为 nil")
	}
}

// Rotate 测试 - AccessTokenExpires 时间映射正确性
func TestRotate_AccessTokenExpiresMapping(t *testing.T) {
	expectedUnix := time.Now().Add(15 * time.Minute).Unix()
	backend := &mockBackend{
		rotateFn: func(ctx context.Context, r *RotateRequest) (*RotateReply, error) {
			return &RotateReply{
				Success:            true,
				AccessToken:        "at",
				AccessTokenExpires: expectedUnix,
				RefreshToken:       "rt",
			}, nil
		},
	}
	dt := NewDualToken(backend)

	resp, err := dt.Rotate(context.Background(), &RefreshParam{
		UserId:                 "user1",
		SessionId:              "sess1",
		RequestRefreshTokenId:  "old",
		NewRefreshTokenId:      "new",
		NewRefreshToken:        "rt",
		NewRefreshTokenExpires: time.Now().Add(time.Hour),
		NewAccessToken:         "at",
		NewAccessTokenExpires:  time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		t.Fatalf("Rotate 不应返回错误: %v", err)
	}
	if resp.AccessTokenExpires.Unix() != expectedUnix {
		t.Fatalf("AccessTokenExpires 期望 unix=%d, 实际 %d", expectedUnix, resp.AccessTokenExpires.Unix())
	}
}

// Rotate 测试 - 过期时间已过的场景, NewRefreshTokenTTL应至少为1s
func TestRotate_ExpiredRefreshTTL(t *testing.T) {
	backend := &mockBackend{
		rotateFn: func(ctx context.Context, r *RotateRequest) (*RotateReply, error) {
			return &RotateReply{Success: true}, nil
		},
	}
	dt := NewDualToken(backend)

	_, err := dt.Rotate(context.Background(), &RefreshParam{
		UserId:                 "user1",
		SessionId:              "sess1",
		RequestRefreshTokenId:  "old",
		NewRefreshTokenId:      "new",
		NewRefreshToken:        "rt",
		NewRefreshTokenExpires: time.Now().Add(-time.Hour), // 已过期
		NewAccessToken:         "at",
		NewAccessTokenExpires:  time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		t.Fatalf("Rotate 不应返回错误: %v", err)
	}

	req := backend.rotateCalls[0]
	if req.NewRefreshTokenTTL != time.Second {
		t.Fatalf("NewRefreshTokenTTL 期望 1s (最小值), 实际 %s", req.NewRefreshTokenTTL)
	}
}

// Revoke 测试 - 正常吊销
func TestRevoke(t *testing.T) {
	backend := &mockBackend{}
	dt := NewDualToken(backend)

	err := dt.Revoke(context.Background(), &RevokeParam{
		UserId:    "user1",
		SessionId: "sess1",
	})
	if err != nil {
		t.Fatalf("Revoke 不应返回错误: %v", err)
	}

	if len(backend.revokeCalls) != 1 {
		t.Fatalf("期望调用 Revoke 1次, 实际 %d次", len(backend.revokeCalls))
	}
	req := backend.revokeCalls[0]
	if req.Key != "voucher:rt:user1:sess1" {
		t.Fatalf("Key 期望 'voucher:rt:user1:sess1', 实际 '%s'", req.Key)
	}
	if req.GraceKey != "voucher:rt:user1:sess1:grace" {
		t.Fatalf("GraceKey 期望 'voucher:rt:user1:sess1:grace', 实际 '%s'", req.GraceKey)
	}
}

// Revoke 测试 - 后端返回错误
func TestRevoke_Error(t *testing.T) {
	wantErr := errors.New("revoke failed")
	backend := &mockBackend{
		revokeFn: func(ctx context.Context, r *RevokeRequest) error {
			return wantErr
		},
	}
	dt := NewDualToken(backend)

	err := dt.Revoke(context.Background(), &RevokeParam{
		UserId:    "user1",
		SessionId: "sess1",
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("期望错误 %v, 实际 %v", wantErr, err)
	}
}

// formatKey 测试 - 自定义前缀
func TestFormatKey_CustomPrefix(t *testing.T) {
	backend := &mockBackend{}
	dt := NewDualToken(backend)
	dt.SetKeyPrefix("my:prefix:")

	// 通过 Save 间接验证 formatKey
	_ = dt.Save(context.Background(), &SaveParam{
		UserId:         "u1",
		SessionId:      "s1",
		RefreshTokenId: "rt",
		Expires:        time.Now().Add(time.Hour),
	})

	if backend.saveCalls[0].Key != "my:prefix:u1:s1" {
		t.Fatalf("Key 期望 'my:prefix:u1:s1', 实际 '%s'", backend.saveCalls[0].Key)
	}
}

// formatResueKey 测试 - 自定义前缀
func TestFormatReuseKey_CustomPrefix(t *testing.T) {
	backend := &mockBackend{
		rotateFn: func(ctx context.Context, r *RotateRequest) (*RotateReply, error) {
			return &RotateReply{Success: true}, nil
		},
	}
	dt := NewDualToken(backend)
	dt.SetKeyPrefix("my:prefix:")

	// 通过 Rotate 间接验证 formatResueKey
	_, _ = dt.Rotate(context.Background(), &RefreshParam{
		UserId:                 "u1",
		SessionId:              "s1",
		RequestRefreshTokenId:  "old",
		NewRefreshTokenId:      "new",
		NewRefreshToken:        "rt",
		NewRefreshTokenExpires: time.Now().Add(time.Hour),
		NewAccessToken:         "at",
		NewAccessTokenExpires:  time.Now().Add(15 * time.Minute),
	})

	if backend.rotateCalls[0].GraceKey != "my:prefix:u1:s1:grace" {
		t.Fatalf("GraceKey 期望 'my:prefix:u1:s1:grace', 实际 '%s'", backend.rotateCalls[0].GraceKey)
	}
}
