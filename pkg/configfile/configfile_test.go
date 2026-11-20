package configfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfigFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create dir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

func setTestHome(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir) // Windows 上 os.UserHomeDir 读取该变量
}

func TestLoad_ExplicitPath(t *testing.T) {
	t.Setenv("KMSCLI_CONFIG_PATH", "")
	explicitPath := filepath.Join(t.TempDir(), "custom.json")
	writeConfigFile(t, explicitPath, `{"authMethod":"ClientKey","regionId":"cn-hangzhou"}`)
	t.Setenv("KMSCLI_CONFIG_PATH", explicitPath)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config from KMSCLI_CONFIG_PATH")
	}
	if cfg.AuthMethod != "ClientKey" || cfg.RegionId != "cn-hangzhou" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestLoad_ExplicitPathMissing(t *testing.T) {
	t.Setenv("KMSCLI_CONFIG_PATH", "")
	missingPath := filepath.Join(t.TempDir(), "missing.json")
	t.Setenv("KMSCLI_CONFIG_PATH", missingPath)

	cfg, err := Load()
	if err == nil {
		t.Fatal("expected error for missing explicit path, got nil")
	}
	if cfg != nil {
		t.Fatalf("expected nil config for missing explicit path, got %+v", cfg)
	}
	if !strings.Contains(err.Error(), missingPath) {
		t.Fatalf("expected error to include path %q, got %q", missingPath, err.Error())
	}
}

func TestLoad_ExplicitPathInvalidJSON(t *testing.T) {
	t.Setenv("KMSCLI_CONFIG_PATH", "")
	invalidPath := filepath.Join(t.TempDir(), "invalid.json")
	writeConfigFile(t, invalidPath, `{not-json`)
	t.Setenv("KMSCLI_CONFIG_PATH", invalidPath)

	cfg, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid JSON at explicit path, got nil")
	}
	if cfg != nil {
		t.Fatalf("expected nil config for invalid JSON at explicit path, got %+v", cfg)
	}
	if !strings.Contains(err.Error(), invalidPath) {
		t.Fatalf("expected error to include path %q, got %q", invalidPath, err.Error())
	}
}

func TestLoad_DefaultPathTakesPriorityOverLegacy(t *testing.T) {
	t.Setenv("KMSCLI_CONFIG_PATH", "")
	home := t.TempDir()
	setTestHome(t, home)

	writeConfigFile(t, filepath.Join(home, ".config", "kmscli", "config.json"), `{"regionId":"default"}`)
	writeConfigFile(t, filepath.Join(home, ".openclaw", "kmscli.json"), `{"regionId":"legacy"}`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg == nil || cfg.RegionId != "default" {
		t.Fatalf("expected default path config, got %+v", cfg)
	}
}

func TestLoad_LegacyFallback(t *testing.T) {
	t.Setenv("KMSCLI_CONFIG_PATH", "")
	home := t.TempDir()
	setTestHome(t, home)

	writeConfigFile(t, filepath.Join(home, ".openclaw", "kmscli.json"), `{"regionId":"legacy"}`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg == nil || cfg.RegionId != "legacy" {
		t.Fatalf("expected legacy config, got %+v", cfg)
	}
}

func TestLoad_InvalidDefaultFallsBackToLegacy(t *testing.T) {
	t.Setenv("KMSCLI_CONFIG_PATH", "")
	home := t.TempDir()
	setTestHome(t, home)

	writeConfigFile(t, filepath.Join(home, ".config", "kmscli", "config.json"), `{not-json`)
	writeConfigFile(t, filepath.Join(home, ".openclaw", "kmscli.json"), `{"regionId":"legacy"}`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg == nil || cfg.RegionId != "legacy" {
		t.Fatalf("expected legacy config after invalid default, got %+v", cfg)
	}
}
