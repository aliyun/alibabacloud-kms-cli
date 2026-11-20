package constants

const (
	EnvKmsEndpoint           = "KMSCLI_KMS_ENDPOINT"
	EnvAuthMethod            = "KMSCLI_AUTH_METHOD"
	EnvAapArn                = "KMSCLI_AAP_ARN"
	EnvTokenPath             = "KMSCLI_TOKEN_PATH"
	EnvClientKeyConfigPath   = "KMSCLI_CLIENT_KEY_CONFIG_PATH"
	EnvClientKeyPasswordPath = "KMSCLI_CLIENT_KEY_PASSWORD_PATH"
	EnvRegionID              = "REGION_ID"
	EnvAlibabaCloudRegionID  = "ALIBABA_CLOUD_REGION_ID"
	EnvEndpointType          = "ENDPOINT_TYPE"

	// DefaultACKTokenPath ACK OIDC 默认 Token 路径
	DefaultACKTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"

	// InstanceGatewayDomainSuffix 专属网关域名后缀
	InstanceGatewayDomainSuffix = "cryptoservice.kms.aliyuncs.com"
)
