package exchanger

import (
	"fmt"
	"strings"

	"github.com/aliyun/alibabacloud-kms-cli/pkg/constants"
)

type Config struct {
	ExchangerType string
	// kms
	KmsEndpoint string
	AapArn      string
	Ca          string
}

const (
	ExchangerTypeKms = "kms"
)

// ValidateConfig 验证配置参数是否合法
func ValidateConfig(config Config) error {
	switch config.ExchangerType {
	case ExchangerTypeKms:
		// ECS 实例认证需要 AAP ID
		if config.KmsEndpoint == "" {
			return fmt.Errorf("KmsEndpoint is required")
		}
		if strings.HasSuffix(config.KmsEndpoint, constants.InstanceGatewayDomainSuffix) {
			if config.Ca == "" {
				return fmt.Errorf("use dedicate endpoint Ca is required")
			}
		}
	default:
		return fmt.Errorf("unsupported exchanger type: %s", config.ExchangerType)
	}
	return nil
}
