package authenticator

import (
	"context"
	"fmt"

	"github.com/aliyun/alibabacloud-kms-cli/pkg/constants"
)

type Identity struct {
	IdentityType  string
	IdentityValue string
}

type Authenticator interface {
	GetIdentity(ctx context.Context) (*Identity, error)
}

// NewAuthenticatorManager 根据配置创建认证管理器
func NewAuthenticatorManager(config Config) (Authenticator, error) {
	// 首先验证配置
	if err := config.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var authenticator Authenticator

	switch config.AuthMethod {
	case ACKOidcJwt:
		tokenPath := config.TokenPath
		// 默认 ack token path
		if tokenPath == "" {
			tokenPath = constants.DefaultACKTokenPath
		}
		authenticator = &AckOIDCAuthenticator{
			TokenPath: tokenPath,
		}
	case ECSInstanceIdentity:
		authenticator = &ECSAuthenticator{}
	default:
		return nil, fmt.Errorf("unsupported authentication method: %s", config.AuthMethod)
	}

	// 使用CachedAuthenticator包装原始认证器，提供缓存功能
	return authenticator, nil
}
