package mauth

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/aliyun/alibabacloud-kms-cli/pkg/mauth/config"
)

// 测试用的 PKCS#12 数据（password = testpassword123）。
// 该证书为测试专用自签名证书（CN=test-client-key，RSA 2048），有效期 10 年，
// 由 openssl 生成，仅用于单元测试，请勿用于任何生产环境。
const testPrivateKeyData = "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"

func prepareClientKeyFiles(t *testing.T) (configPath, passwordPath string) {
	tmpDir := t.TempDir()
	configPath = filepath.Join(tmpDir, "client_key.json")
	passwordPath = filepath.Join(tmpDir, "password.txt")

	clientKeyConfig := map[string]string{
		"KeyId":          "test-key-id-12345",
		"PrivateKeyData": testPrivateKeyData,
	}
	data, err := json.Marshal(clientKeyConfig)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(passwordPath, []byte("testpassword123"), 0644); err != nil {
		t.Fatal(err)
	}
	return configPath, passwordPath
}

func TestNewMAuth_ClientKey(t *testing.T) {
	configPath, passwordPath := prepareClientKeyFiles(t)
	c := config.AuthConfig{
		AuthMethod:            config.ClientKey,
		KmsEndpoint:           "https://kms.cn-hangzhou.aliyuncs.com",
		ClientKeyConfigPath:   configPath,
		ClientKeyPasswordPath: passwordPath,
	}

	mAuth, err := NewMAuth(c, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mAuth == nil {
		t.Fatal("expected non-nil MAuth")
	}
	if mAuth.config.AuthMethod != config.ClientKey {
		t.Fatalf("expected ClientKey auth method")
	}
}

func TestMAuth_GetToken_ClientKey(t *testing.T) {
	configPath, passwordPath := prepareClientKeyFiles(t)
	c := config.AuthConfig{
		AuthMethod:            config.ClientKey,
		KmsEndpoint:           "https://kms.cn-hangzhou.aliyuncs.com",
		ClientKeyConfigPath:   configPath,
		ClientKeyPasswordPath: passwordPath,
	}

	mAuth, err := NewMAuth(c, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	token, err := mAuth.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}
	if token == nil {
		t.Fatal("expected non-nil token")
	}
	if token.TokenType != AuthTokenTypeClientKey {
		t.Fatalf("expected token type %s, got %s", AuthTokenTypeClientKey, token.TokenType)
	}
	if token.TokenKeyId != "test-key-id-12345" {
		t.Fatalf("expected key id test-key-id-12345, got %s", token.TokenKeyId)
	}
	if token.TokenValue == "" {
		t.Fatal("expected non-empty token value")
	}
	if time.Until(token.ExpiresAt) < 24*time.Hour {
		t.Fatalf("expected token to be valid for more than 24 hours, got %v", time.Until(token.ExpiresAt))
	}
}

func TestMAuth_GetToken_ClientKey_Cache(t *testing.T) {
	configPath, passwordPath := prepareClientKeyFiles(t)
	c := config.AuthConfig{
		AuthMethod:            config.ClientKey,
		KmsEndpoint:           "https://kms.cn-hangzhou.aliyuncs.com",
		ClientKeyConfigPath:   configPath,
		ClientKeyPasswordPath: passwordPath,
	}

	mAuth, err := NewMAuth(c, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 第一次获取
	token1, err := mAuth.GetToken(context.Background())
	if err != nil {
		t.Fatalf("first GetToken failed: %v", err)
	}

	// 第二次获取（应该从缓存返回）
	token2, err := mAuth.GetToken(context.Background())
	if err != nil {
		t.Fatalf("second GetToken failed: %v", err)
	}

	// 两个 token 应该是同一个对象（缓存命中）
	if token1 != token2 {
		t.Fatal("expected cached token to be returned")
	}
}

func TestMAuth_ClearToken_ClientKey(t *testing.T) {
	configPath, passwordPath := prepareClientKeyFiles(t)
	c := config.AuthConfig{
		AuthMethod:            config.ClientKey,
		KmsEndpoint:           "https://kms.cn-hangzhou.aliyuncs.com",
		ClientKeyConfigPath:   configPath,
		ClientKeyPasswordPath: passwordPath,
	}

	mAuth, err := NewMAuth(c, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 先获取 token 填充缓存
	_, err = mAuth.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	// 清除缓存
	if err := mAuth.ClearToken(context.Background()); err != nil {
		t.Fatalf("ClearToken failed: %v", err)
	}

	// 再次获取，应该是新的 token（因为缓存被清除，重新读取文件）
	token2, err := mAuth.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken after clear failed: %v", err)
	}
	if token2 == nil {
		t.Fatal("expected non-nil token after clear")
	}
}

func TestMAuth_GetToken_ClientKey_WrongPassword(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "client_key.json")
	passwordPath := filepath.Join(tmpDir, "password.txt")

	clientKeyConfig := map[string]string{
		"KeyId":          "test-key-id",
		"PrivateKeyData": testPrivateKeyData,
	}
	data, _ := json.Marshal(clientKeyConfig)
	os.WriteFile(configPath, data, 0644)
	os.WriteFile(passwordPath, []byte("wrongpassword"), 0644)

	c := config.AuthConfig{
		AuthMethod:            config.ClientKey,
		KmsEndpoint:           "https://kms.cn-hangzhou.aliyuncs.com",
		ClientKeyConfigPath:   configPath,
		ClientKeyPasswordPath: passwordPath,
	}

	mAuth, err := NewMAuth(c, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = mAuth.GetToken(context.Background())
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

// TestMAuth_ConcurrentTokenAccess 是 RefreshToken 锁边界的并发回归测试。
// 预置缓存后，多组 goroutine 并发执行 GetToken、ClearToken 与 RefreshToken，
// 修复前 RefreshToken 在锁外读取 m.cacheToken 状态，-race 下应能复现竞态。
func TestMAuth_ConcurrentTokenAccess(t *testing.T) {
	configPath, passwordPath := prepareClientKeyFiles(t)
	c := config.AuthConfig{
		AuthMethod:            config.ClientKey,
		KmsEndpoint:           "https://kms.cn-hangzhou.aliyuncs.com",
		ClientKeyConfigPath:   configPath,
		ClientKeyPasswordPath: passwordPath,
	}

	mAuth, err := NewMAuth(c, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 预置缓存
	if _, err := mAuth.GetToken(context.Background()); err != nil {
		t.Fatalf("prepopulate GetToken failed: %v", err)
	}

	const (
		workers    = 9
		iterations = 200
	)
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				switch worker % 3 {
				case 0:
					if _, err := mAuth.GetToken(ctx); err != nil {
						t.Errorf("worker %d GetToken failed: %v", worker, err)
					}
				case 1:
					if err := mAuth.ClearToken(ctx); err != nil {
						t.Errorf("worker %d ClearToken failed: %v", worker, err)
					}
				case 2:
					mAuth.RefreshToken()
				}
			}
		}(i)
	}
	wg.Wait()
}
