package utils

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// 使用 "none" alg 构造的测试 JWT，payload 包含 exp=1893456000 (2030-01-01T00:00:00Z)
const testJWT = "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJleHAiOjE4OTM0NTYwMDAsInN1YiI6InRlc3QifQ."

func TestIsValidJWT_Valid(t *testing.T) {
	if !IsValidJWT(testJWT) {
		t.Fatal("expected valid JWT")
	}
}

func TestIsValidJWT_Invalid(t *testing.T) {
	invalidTokens := []string{
		"not.a.jwt",
		"only.two.parts",
		"",
		"a.b.c", // c 可以是任意字符串，但 a,b 必须是合法 base64url JSON
	}
	for _, token := range invalidTokens {
		if IsValidJWT(token) {
			t.Fatalf("expected invalid JWT for: %s", token)
		}
	}
}

func TestParseJWTClaims_Success(t *testing.T) {
	claims, err := ParseJWTClaims(testJWT)
	if err != nil {
		t.Fatalf("ParseJWTClaims failed: %v", err)
	}
	if claims["sub"] != "test" {
		t.Fatalf("expected sub=test, got: %v", claims["sub"])
	}
}

func TestGetJWTExpiration_Success(t *testing.T) {
	exp, err := GetJWTExpiration(testJWT)
	if err != nil {
		t.Fatalf("GetJWTExpiration failed: %v", err)
	}
	expected := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	if !exp.Equal(expected) {
		t.Fatalf("expected expiration %v, got: %v", expected, exp)
	}
}

func TestGetJWTExpiration_MissingExp(t *testing.T) {
	// header={"alg":"none"}, payload={"sub":"test"}
	jwtNoExp := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ0ZXN0In0."
	_, err := GetJWTExpiration(jwtNoExp)
	if err == nil {
		t.Fatal("expected error when exp is missing")
	}
}

func TestIsValidJWTFile(t *testing.T) {
	tmpDir := t.TempDir()

	// 测试存在的有效 JWT 文件
	validPath := filepath.Join(tmpDir, "valid.jwt")
	if err := os.WriteFile(validPath, []byte(testJWT), 0644); err != nil {
		t.Fatal(err)
	}
	if !IsValidJWTFile(validPath) {
		t.Fatal("expected valid JWT file")
	}

	// 测试不存在的文件
	if IsValidJWTFile(filepath.Join(tmpDir, "notexist.jwt")) {
		t.Fatal("expected invalid for non-existent file")
	}

	// 测试存在的无效 JWT 文件
	invalidPath := filepath.Join(tmpDir, "invalid.jwt")
	if err := os.WriteFile(invalidPath, []byte("not.a.jwt"), 0644); err != nil {
		t.Fatal(err)
	}
	if IsValidJWTFile(invalidPath) {
		t.Fatal("expected invalid JWT file")
	}
}
