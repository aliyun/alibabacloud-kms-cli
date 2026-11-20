package authenticator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aliyun/alibabacloud-kms-cli/pkg/constants"
)

func TestNewAuthenticatorManager_AckOIDC(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.jwt")
	os.WriteFile(tokenPath, []byte("eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ0ZXN0In0."), 0644)
	auth, err := NewAuthenticatorManager(Config{
		AuthMethod: ACKOidcJwt,
		TokenPath:  tokenPath,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if auth == nil {
		t.Fatal("expected non-nil authenticator")
	}
	_, ok := auth.(*AckOIDCAuthenticator)
	if !ok {
		t.Fatalf("expected *AckOIDCAuthenticator, got %T", auth)
	}
}

func TestNewAuthenticatorManager_ECS(t *testing.T) {
	auth, err := NewAuthenticatorManager(Config{
		AuthMethod: ECSInstanceIdentity,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if auth == nil {
		t.Fatal("expected non-nil authenticator")
	}
	_, ok := auth.(*ECSAuthenticator)
	if !ok {
		t.Fatalf("expected *ECSAuthenticator, got %T", auth)
	}
}

func TestNewAuthenticatorManager_DefaultTokenPath(t *testing.T) {
	auth, err := NewAuthenticatorManager(Config{
		AuthMethod: ACKOidcJwt,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ackAuth, ok := auth.(*AckOIDCAuthenticator)
	if !ok {
		t.Fatalf("expected *AckOIDCAuthenticator, got %T", auth)
	}
	if ackAuth.TokenPath != constants.DefaultACKTokenPath {
		t.Fatalf("expected default token path %s, got %s", constants.DefaultACKTokenPath, ackAuth.TokenPath)
	}
}

func TestNewAuthenticatorManager_UnsupportedMethod(t *testing.T) {
	_, err := NewAuthenticatorManager(Config{
		AuthMethod: "UnknownMethod",
	})
	if err == nil {
		t.Fatal("expected error for unsupported method")
	}
}

func TestNewAuthenticatorManager_InvalidConfig(t *testing.T) {
	_, err := NewAuthenticatorManager(Config{})
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}
