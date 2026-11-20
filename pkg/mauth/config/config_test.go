package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAuthConfigValidate_EmptyAuthMethod(t *testing.T) {
	c := AuthConfig{KmsEndpoint: "https://kms.cn-hangzhou.aliyuncs.com"}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for empty AuthMethod")
	}
}

func TestAuthConfigValidate_EmptyKmsEndpoint(t *testing.T) {
	c := AuthConfig{AuthMethod: ECSInstanceIdentity, AapArn: "arn:test"}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for empty KmsEndpoint")
	}
}

func TestAuthConfigValidate_ECSMissingAapArn(t *testing.T) {
	c := AuthConfig{
		AuthMethod:  ECSInstanceIdentity,
		KmsEndpoint: "https://kms.cn-hangzhou.aliyuncs.com",
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for missing AapArn")
	}
}

func TestAuthConfigValidate_ACKMissingAapArn(t *testing.T) {
	c := AuthConfig{
		AuthMethod:  ACKOidcJwt,
		KmsEndpoint: "https://kms.cn-hangzhou.aliyuncs.com",
		TokenPath:   "/var/run/secrets/tokens/token",
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for missing AapArn")
	}
}

func TestAuthConfigValidate_ClientKeyMissingPaths(t *testing.T) {
	c := AuthConfig{
		AuthMethod:  ClientKey,
		KmsEndpoint: "https://kms.cn-hangzhou.aliyuncs.com",
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for missing ClientKey paths")
	}
}

func TestAuthConfigValidate_ClientKeyMissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	c := AuthConfig{
		AuthMethod:            ClientKey,
		KmsEndpoint:           "https://kms.cn-hangzhou.aliyuncs.com",
		ClientKeyConfigPath:   filepath.Join(tmpDir, "notexist.json"),
		ClientKeyPasswordPath: filepath.Join(tmpDir, "notexist.txt"),
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for non-existent config file")
	}
}

func TestAuthConfigValidate_ClientKeySuccess(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "client_key.json")
	passwordPath := filepath.Join(tmpDir, "password.txt")
	os.WriteFile(configPath, []byte(`{"KeyId":"test","PrivateKeyData":"data"}`), 0644)
	os.WriteFile(passwordPath, []byte("password"), 0644)

	c := AuthConfig{
		AuthMethod:            ClientKey,
		KmsEndpoint:           "https://kms.cn-hangzhou.aliyuncs.com",
		ClientKeyConfigPath:   configPath,
		ClientKeyPasswordPath: passwordPath,
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthConfigValidate_ECSSuccess(t *testing.T) {
	c := AuthConfig{
		AuthMethod:  ECSInstanceIdentity,
		KmsEndpoint: "https://kms.cn-hangzhou.aliyuncs.com",
		AapArn:      "arn:test",
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthConfigValidate_UnsupportedMethod(t *testing.T) {
	c := AuthConfig{
		AuthMethod:  "UnknownMethod",
		KmsEndpoint: "https://kms.cn-hangzhou.aliyuncs.com",
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for unsupported auth method")
	}
}

func TestInferAuthMethod_ClientKey(t *testing.T) {
	c := AuthConfig{ClientKeyConfigPath: "/path/to/key"}
	c.InferAuthMethod()
	if c.AuthMethod != ClientKey {
		t.Fatalf("expected %s, got %s", ClientKey, c.AuthMethod)
	}
}

func TestInferAuthMethod_ACKOidcJwt(t *testing.T) {
	c := AuthConfig{TokenPath: "/var/run/secrets/tokens/token"}
	c.InferAuthMethod()
	if c.AuthMethod != ACKOidcJwt {
		t.Fatalf("expected %s, got %s", ACKOidcJwt, c.AuthMethod)
	}
}

func TestInferAuthMethod_ECSInstanceIdentity(t *testing.T) {
	c := AuthConfig{AapArn: "arn:acs:kms:cn-hangzhou:123456789:aap/test"}
	c.InferAuthMethod()
	if c.AuthMethod != ECSInstanceIdentity {
		t.Fatalf("expected %s, got %s", ECSInstanceIdentity, c.AuthMethod)
	}
}

func TestInferAuthMethod_Default(t *testing.T) {
	c := AuthConfig{}
	c.InferAuthMethod()
	if c.AuthMethod != "" {
		t.Fatalf("expected empty, got %s", c.AuthMethod)
	}
}

func TestInferAuthMethod_PresetNotOverridden(t *testing.T) {
	c := AuthConfig{AuthMethod: ClientKey, TokenPath: "/var/run/secrets/tokens/token"}
	c.InferAuthMethod()
	if c.AuthMethod != ClientKey {
		t.Fatalf("expected preset %s not to be overridden, got %s", ClientKey, c.AuthMethod)
	}
}

func TestInferAuthMethod_Priority_ClientKeyOverToken(t *testing.T) {
	c := AuthConfig{ClientKeyConfigPath: "/path/to/key", TokenPath: "/var/run/secrets/tokens/token"}
	c.InferAuthMethod()
	if c.AuthMethod != ClientKey {
		t.Fatalf("expected ClientKey to take priority over ACKOidcJwt, got %s", c.AuthMethod)
	}
}

func TestInferAuthMethod_Priority_TokenOverAapArn(t *testing.T) {
	c := AuthConfig{TokenPath: "/var/run/secrets/tokens/token", AapArn: "arn:test"}
	c.InferAuthMethod()
	if c.AuthMethod != ACKOidcJwt {
		t.Fatalf("expected ACKOidcJwt to take priority over ECSInstanceIdentity, got %s", c.AuthMethod)
	}
}
