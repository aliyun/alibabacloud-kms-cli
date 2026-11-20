package common

import (
	"os"

	"github.com/aliyun/alibabacloud-kms-cli/pkg/configfile"
	"github.com/aliyun/alibabacloud-kms-cli/pkg/constants"
	"github.com/aliyun/alibabacloud-kms-cli/pkg/mauth/config"
	"github.com/spf13/cobra"
)

// RegisterKmsFlags 向指定命令注册 KMS 相关的通用 flag
// 包括 endpoint、认证方式、aap-arn、token-path、client-key 等
func RegisterKmsFlags(cmd *cobra.Command, kmsEndpoint, authMethod, aapArn, tokenPath, clientKeyConfigPath, clientKeyPasswordPath *string) {
	cmd.Flags().StringVarP(kmsEndpoint, "kms-endpoint", "e", os.Getenv(constants.EnvKmsEndpoint), "KMS endpoint")
	cmd.Flags().StringVarP(authMethod, "auth-method", "A", os.Getenv(constants.EnvAuthMethod), "Authentication method (ECSInstanceIdentity, ACKOidcJwt, ClientKey)")
	cmd.Flags().StringVarP(aapArn, "aap-arn", "a", os.Getenv(constants.EnvAapArn), "AAP ARN")
	cmd.Flags().StringVarP(tokenPath, "token-path", "t", os.Getenv(constants.EnvTokenPath), "Token path for ACK OIDC")
	cmd.Flags().StringVarP(clientKeyConfigPath, "client-key-config-path", "k", os.Getenv(constants.EnvClientKeyConfigPath), "ClientKey config file path")
	cmd.Flags().StringVarP(clientKeyPasswordPath, "client-key-password-path", "p", os.Getenv(constants.EnvClientKeyPasswordPath), "ClientKey password file path")
}

// MergeConfigFile 将配置文件中的值合并到 AuthConfig 中。
// 只有当 AuthConfig 对应字段为空时，才会使用配置文件中的值填充。
// 同时会将配置文件的 regionId 和 endpointType 写入对应环境变量（如果环境变量未设置）。
func MergeConfigFile(cfgFile *configfile.Config, authConfig *config.AuthConfig) {
	if cfgFile == nil {
		return
	}
	if authConfig.AuthMethod == "" {
		authConfig.AuthMethod = cfgFile.AuthMethod
	}
	if authConfig.KmsEndpoint == "" {
		authConfig.KmsEndpoint = cfgFile.KmsEndpoint
	}
	if authConfig.AapArn == "" {
		authConfig.AapArn = cfgFile.AapArn
	}
	if authConfig.TokenPath == "" {
		authConfig.TokenPath = cfgFile.TokenPath
	}
	if authConfig.ClientKeyConfigPath == "" {
		authConfig.ClientKeyConfigPath = cfgFile.ClientKeyConfigPath
	}
	if authConfig.ClientKeyPasswordPath == "" {
		authConfig.ClientKeyPasswordPath = cfgFile.ClientKeyPasswordPath
	}
	if cfgFile.RegionId != "" && os.Getenv(constants.EnvRegionID) == "" {
		os.Setenv(constants.EnvRegionID, cfgFile.RegionId)
	}
	if cfgFile.EndpointType != "" && os.Getenv(constants.EnvEndpointType) == "" {
		os.Setenv(constants.EnvEndpointType, cfgFile.EndpointType)
	}
}
