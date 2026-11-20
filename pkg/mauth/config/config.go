package config

import (
	"fmt"
	"os"

	"github.com/aliyun/alibabacloud-kms-cli/pkg/constants"
)

// AuthConfig 认证配置
type AuthConfig struct {
	// 认证方式
	AuthMethod string `json:"AuthMethod"`
	// AAP ID
	AapArn string `json:"AapArn"`
	// 令牌路径（用于ACK OIDC）
	TokenPath string `json:"TokenPath,omitempty"`
	// ClientKey配置路径
	ClientKeyConfigPath string `json:"ClientKeyConfigPath,omitempty"`
	// ClientKey Password
	ClientKeyPasswordPath string `json:"ClientKeyPasswordPath,omitempty"`
	// KmsEndpoint KMS端点
	KmsEndpoint string `json:"KmsEndpoint,omitempty"`
	Ca          string `json:"Ca,omitempty"`
}

const (
	ECSInstanceIdentity = "ECSInstanceIdentity"
	ACKOidcJwt          = "ACKOidcJwt"
	ClientKey           = "ClientKey"
)

// InferAuthMethod 根据其他参数自动推断认证方式
func (c *AuthConfig) InferAuthMethod() {
	if c.AuthMethod != "" {
		return
	}
	if c.ClientKeyConfigPath != "" || c.ClientKeyPasswordPath != "" {
		c.AuthMethod = ClientKey
		return
	}
	if c.TokenPath != "" {
		c.AuthMethod = ACKOidcJwt
		return
	}
	// 仅在同时存在默认 ACK token 文件和 AapArn 时才推断为 ACKOidcJwt。
	// 避免在容器环境（文件存在但 AapArn 为空）中错误推断，导致想使用默认凭证链的用户启动失败。
	if _, err := os.Stat(constants.DefaultACKTokenPath); err == nil {
		if c.AapArn != "" {
			c.AuthMethod = ACKOidcJwt
			c.TokenPath = constants.DefaultACKTokenPath
			return
		}
	}
	if c.AapArn != "" {
		c.AuthMethod = ECSInstanceIdentity
		return
	}
}

// Validate 校验配置参数
func (c *AuthConfig) Validate() error {
	// 检查认证方式是否为空
	if c.AuthMethod == "" {
		return fmt.Errorf("AuthMethod is required")
	}
	if c.KmsEndpoint == "" {
		return fmt.Errorf("KmsEndpoint is required")
	}

	// 检查不同认证方式的特定要求
	switch c.AuthMethod {
	case ACKOidcJwt, ECSInstanceIdentity:
		if c.AapArn == "" {
			return fmt.Errorf("AapArn is required for %s authentication", c.AuthMethod)
		}
	case ClientKey:
		// Client Key 认证不需要 AAP ID
		// ClientKeyConfigPath 不为空时，需要校验文件是否存在
		if c.ClientKeyPasswordPath == "" || c.ClientKeyConfigPath == "" {
			return fmt.Errorf("ClientKeyConfigPath and ClientKeyPasswordPath are required for Client Key authentication")
		}
		if _, err := os.Stat(c.ClientKeyConfigPath); os.IsNotExist(err) {
			return fmt.Errorf("client key config file does not exist: %s", c.ClientKeyConfigPath)
		}
		if _, err := os.Stat(c.ClientKeyPasswordPath); os.IsNotExist(err) {
			return fmt.Errorf("client key password file does not exist: %s", c.ClientKeyPasswordPath)
		}
	default:
		return fmt.Errorf("unsupported authentication method: %s", c.AuthMethod)
	}

	return nil
}
