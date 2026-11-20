package configfile

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

// Config 配置文件结构
type Config struct {
	AuthMethod            string `json:"authMethod"`
	AapArn                string `json:"aapArn"`
	TokenPath             string `json:"tokenPath"`
	ClientKeyConfigPath   string `json:"clientKeyConfigPath"`
	ClientKeyPasswordPath string `json:"clientKeyPasswordPath"`
	KmsEndpoint           string `json:"kmsEndpoint"`
	RegionId              string `json:"regionId"`
	EndpointType          string `json:"endpointType"`
}

// Load 加载配置文件。
// 优先从 KMSCLI_CONFIG_PATH 环境变量读取路径，否则使用默认路径 ~/.config/kmscli/config.json
// 为兼容旧版本，若默认路径不存在、为空或 JSON 解析失败，且 ~/.openclaw/kmscli.json 存在，则读取旧路径。
// KMSCLI_CONFIG_PATH 显式指定的路径加载失败时返回带路径信息的错误；
// 默认路径及 legacy 路径均不可用时返回 nil，不报错。
func Load() (*Config, error) {
	configPath := os.Getenv("KMSCLI_CONFIG_PATH")
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = getHomeDirRobust()
			if homeDir == "" {
				return tryLoadLegacy("")
			}
		}
		configPath = filepath.Join(homeDir, ".config", "kmscli", "config.json")
		// 兼容旧路径：默认路径不存在时才考虑 legacy
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return tryLoadLegacy(configPath)
		}
		// 默认路径存在但读取/解析失败时，也回退到 legacy
		cfg, ok := loadFromPath(configPath)
		if ok {
			return cfg, nil
		}
		return tryLoadLegacy(configPath)
	}

	// KMSCLI_CONFIG_PATH 明确指定了路径，只尝试该路径
	cfg, ok := loadFromPath(configPath)
	if ok {
		return cfg, nil
	}
	return nil, fmt.Errorf("failed to load config file from KMSCLI_CONFIG_PATH %q", configPath)
}

// loadFromPath 从指定路径加载配置，成功返回 (cfg, true)，失败返回 (nil, false)
func loadFromPath(path string) (*Config, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, false
	}
	return &cfg, true
}

// tryLoadLegacy 尝试从旧路径 ~/.openclaw/kmscli.json 加载配置。
// 若 legacy 路径不存在，则回退到当前工作目录下的 kmscli.json 或 .openclaw/kmscli.json。
func tryLoadLegacy(alreadyTried string) (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = getHomeDirRobust()
	}
	if homeDir != "" {
		legacyPath := filepath.Join(homeDir, ".openclaw", "kmscli.json")
		if legacyPath != alreadyTried {
			cfg, ok := loadFromPath(legacyPath)
			if ok {
				return cfg, nil
			}
		}
	}
	// homeDir 为空或 legacy 路径不存在时，回退到当前工作目录
	cwd, _ := os.Getwd()
	if cwd != "" {
		for _, rel := range []string{"kmscli.json", ".openclaw/kmscli.json"} {
			fallbackPath := filepath.Join(cwd, rel)
			if fallbackPath != alreadyTried {
				cfg, ok := loadFromPath(fallbackPath)
				if ok {
					return cfg, nil
				}
			}
		}
	}
	return nil, nil
}

// getHomeDirRobust 在 os.UserHomeDir() 失败时通过环境变量或系统用户表获取 home 目录
func getHomeDirRobust() string {
	// 1. 环境变量 HOME
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	// 2. 通过系统用户表获取
	if current, err := user.Current(); err == nil {
		return current.HomeDir
	}
	return ""
}
