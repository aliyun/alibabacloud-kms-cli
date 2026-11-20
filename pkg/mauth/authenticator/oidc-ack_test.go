package authenticator

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestAckOIDCAuthenticator_GetIdentity_Success(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.jwt")
	// 使用测试 JWT（header={"alg":"none","typ":"JWT"}, payload={"sub":"test"}）
	token := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ0ZXN0In0."
	if err := os.WriteFile(tokenPath, []byte(token), 0644); err != nil {
		t.Fatal(err)
	}

	auth := &AckOIDCAuthenticator{TokenPath: tokenPath}
	identity, err := auth.GetIdentity(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if identity.IdentityType != IdentityTypeJWT {
		t.Fatalf("expected identity type %s, got %s", IdentityTypeJWT, identity.IdentityType)
	}
	if identity.IdentityValue != token {
		t.Fatalf("expected identity value %s, got %s", token, identity.IdentityValue)
	}
}

func TestAckOIDCAuthenticator_GetIdentity_FileNotExist(t *testing.T) {
	auth := &AckOIDCAuthenticator{TokenPath: "/nonexistent/path/token.jwt"}
	_, err := auth.GetIdentity(context.Background())
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestAckOIDCAuthenticator_GetIdentity_InvalidJWT(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "bad-token.txt")
	if err := os.WriteFile(tokenPath, []byte("not-a-jwt"), 0644); err != nil {
		t.Fatal(err)
	}

	auth := &AckOIDCAuthenticator{TokenPath: tokenPath}
	_, err := auth.GetIdentity(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid JWT")
	}
}
