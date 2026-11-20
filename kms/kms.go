package kms

import (
	"context"
	"fmt"
	"github.com/aliyun/alibabacloud-kms-cli/pkg/mauth/utils"
	"math"
	"os"
	"strings"
	"time"

	"github.com/aliyun/alibabacloud-kms-cli/pkg/constants"
	"github.com/aliyun/alibabacloud-kms-cli/pkg/mauth"
	"github.com/aliyun/alibabacloud-kms-cli/pkg/mauth/config"
	"github.com/aliyun/alibabacloud-kms-cli/useragent"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	kms20160120 "github.com/alibabacloud-go/kms-20160120/v3/client"
	teautil "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/credentials-go/credentials"
)

const (
	ActionGetSecretValue        = "GetSecretValue"
	ActionDecrypt               = "Decrypt"
	InvalidAccessKeyIdNotFound  = "InvalidAccessKeyId.NotFound"
	ProtocolHTTPS               = "https"
	DefaultReadTimeoutMs        = 3000
	DefaultConnectTimeoutMs     = 3000
	MaxRetryAttempts            = 3
	InitialBackoffDelayMs       = 200
	CommandGetSecrets           = "getsecret"
	CommandGetSecretsAndDecrypt = "getsecret-decrypt"
)

func CreateKmsClient() (result *kms20160120.Client, err error) {
	// 工程代码建议使用更安全的无AK方式，凭据配置方式请参见：https://help.aliyun.com/document_detail/378661.html。
	credential, err := credentials.NewCredential(nil)
	if err != nil {
		return result, err
	}

	config := &openapi.Config{
		Credential: credential,
	}
	config.UserAgent = tea.String(useragent.GetUserAgent())
	kmsEndpoint, err := ResolveKmsEndpoint("")
	if err != nil {
		return result, err
	}
	config.Endpoint = tea.String(kmsEndpoint)

	return kms20160120.NewClient(config)
}

func GetSecretValue(client *kms20160120.Client, secretName string) (string, error) {
	req := &kms20160120.GetSecretValueRequest{
		SecretName: tea.String(secretName),
	}

	resp, err := client.GetSecretValue(req)
	if err != nil {
		return "", fmt.Errorf("GetSecretValue Error: %v", err)
	}
	return *resp.Body.GetSecretData(), nil
}

func GetSecretValueAndDecrypt(client *kms20160120.Client, secretName string) (string, error) {
	cipherTextBlob, err := GetSecretValue(client, secretName)
	if err != nil {
		return "", fmt.Errorf("GetSecretValue Error: %v", err)
	}

	req := &kms20160120.DecryptRequest{
		CiphertextBlob: tea.String(cipherTextBlob),
	}
	resp, err := client.Decrypt(req)
	if err != nil {
		return "", fmt.Errorf("decrypt Error: %v", err)
	}
	return *resp.Body.Plaintext, nil
}

type SecretManagerClientWithBearer struct {
	mauth           *mauth.MAuth
	commonApiConfig *openapi.Config
}

func NewSecretManagerClientWithBearer(authConfig config.AuthConfig) (*SecretManagerClientWithBearer, error) {
	if authConfig.AuthMethod == config.ACKOidcJwt || authConfig.AuthMethod == config.ClientKey {
		if authConfig.KmsEndpoint == "" && os.Getenv(constants.EnvRegionID) == "" && os.Getenv(constants.EnvAlibabaCloudRegionID) == "" {
			if authConfig.AuthMethod == config.ACKOidcJwt {
				if regionId := extractRegionFromArn(authConfig.AapArn); regionId != "" {
					os.Setenv(constants.EnvRegionID, regionId)
				} else {
					return nil, fmt.Errorf("region id is required for ACKOidcJwt authentication, please set REGION_ID or ALIBABA_CLOUD_REGION_ID environment variable, or specify --kms-endpoint")
				}
			} else {
				return nil, fmt.Errorf("region id is required for %s authentication, please set REGION_ID or ALIBABA_CLOUD_REGION_ID environment variable, or specify --kms-endpoint", authConfig.AuthMethod)
			}
		}
	}

	kmsEndpoint, err := ResolveKmsEndpoint(authConfig.KmsEndpoint)
	if err != nil {
		return nil, err
	}
	authConfig.KmsEndpoint = kmsEndpoint

	// 自动填充内置 CA
	if authConfig.Ca == "" {
		authConfig.Ca = utils.ResolveBuiltinCa(kmsEndpoint, authConfig.AapArn)
	}

	mauthObj, err := mauth.NewMAuth(authConfig, nil)
	if err != nil {
		return nil, fmt.Errorf("new mauth error: %w", err)
	}

	openApiConfig := &openapi.Config{
		Endpoint:       tea.String(kmsEndpoint),
		Protocol:       tea.String(ProtocolHTTPS),
		ReadTimeout:    tea.Int(DefaultReadTimeoutMs),
		ConnectTimeout: tea.Int(DefaultConnectTimeoutMs),
	}
	if authConfig.Ca != "" {
		openApiConfig.Ca = tea.String(authConfig.Ca)
	}

	return &SecretManagerClientWithBearer{
		mauth:           mauthObj,
		commonApiConfig: openApiConfig,
	}, nil
}

// extractRegionFromArn 从 AAP ARN 中提取 regionId。
// ARN 格式：acs:kms:{region}:{account}:aap/{id}
func extractRegionFromArn(arn string) string {
	parts := strings.Split(arn, ":")
	if len(parts) >= 4 {
		return parts[2]
	}
	return ""
}

func (sm *SecretManagerClientWithBearer) GetSecretValue(ctx context.Context, secretName string) (string, error) {
	request := &kms20160120.GetSecretValueRequest{
		SecretName: tea.String(secretName),
	}

	var ret *kms20160120.GetSecretValueResponse
	var err error
	for i := 0; i < MaxRetryAttempts; i++ {
		ret, err = sm.getSecretValue(ctx, request)
		if err == nil {
			break
		}

		sdkErr, ok := err.(*tea.SDKError)
		if !ok {
			break
		}

		if *sdkErr.Code == InvalidAccessKeyIdNotFound {
			clearTokenErr := sm.mauth.ClearToken(ctx)
			if clearTokenErr != nil {
				return "", fmt.Errorf("encountered %s[%v] and clear token error[%v]", InvalidAccessKeyIdNotFound, sdkErr, clearTokenErr)
			}
			continue
		}

		if *sdkErr.StatusCode == 429 || *sdkErr.StatusCode >= 500 {
			time.Sleep(time.Duration(getWaitTimeExponential(i+1) * int64(time.Millisecond)))
			continue
		} else {
			break
		}
	}

	if err != nil {
		return "", err
	}
	return *ret.Body.GetSecretData(), nil
}

func (sm *SecretManagerClientWithBearer) buildSignedClient(ctx context.Context, request interface{}, action string) (*kms20160120.Client, error) {
	token, err := sm.mauth.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("get token error: %w", err)
	}
	if token == nil {
		return nil, fmt.Errorf("get token error: token is nil")
	}

	queryParams, header, err := ExtractSignParams(request, token.TokenKeyId, action)
	if err != nil {
		return nil, fmt.Errorf("extract sign params error: %w", err)
	}

	signature, err := Sign(queryParams, header, token.TokenValue)
	if err != nil {
		return nil, fmt.Errorf("sign error: %w", err)
	}

	newOpenapiConfig := &openapi.Config{
		RegionId:       sm.commonApiConfig.RegionId,
		ReadTimeout:    sm.commonApiConfig.ReadTimeout,
		ConnectTimeout: sm.commonApiConfig.ConnectTimeout,
		Protocol:       sm.commonApiConfig.Protocol,
		Endpoint:       sm.commonApiConfig.Endpoint,
		Ca:             sm.commonApiConfig.Ca,
		BearerToken:    tea.String(signature),
	}
	kmsClient, err := kms20160120.NewClient(newOpenapiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create sm client: %v", err)
	}

	if kmsClient.Headers == nil {
		kmsClient.Headers = map[string]*string{}
	}
	for key, value := range header {
		kmsClient.Headers[key] = tea.String(value)
	}

	return kmsClient, nil
}

func (sm *SecretManagerClientWithBearer) getSecretValue(ctx context.Context, request *kms20160120.GetSecretValueRequest) (*kms20160120.GetSecretValueResponse, error) {
	kmsClient, err := sm.buildSignedClient(ctx, request, ActionGetSecretValue)
	if err != nil {
		return nil, err
	}
	runtime := &teautil.RuntimeOptions{}
	runtime.Autoretry = tea.Bool(true)
	runtime.MaxAttempts = tea.Int(MaxRetryAttempts)
	return kmsClient.GetSecretValueWithOptions(request, runtime)
}

func getWaitTimeExponential(retryCount int) int64 {
	var initialDelay float64 = InitialBackoffDelayMs //ms
	return int64(math.Pow(2, float64(retryCount)) * initialDelay)
}

// ResolveSecret 根据认证配置解析 Secret 值
// 支持 mauth 方式和默认凭证链方式
// mauth 方式仅支持 getsecret，默认凭证链方式支持 getsecret 和 getsecret-decrypt
func ResolveSecret(ctx context.Context, authConfig config.AuthConfig, command, secretName string) (string, error) {
	authConfig.InferAuthMethod()
	if authConfig.AuthMethod != "" {
		if command == CommandGetSecretsAndDecrypt {
			return "", fmt.Errorf("getsecret-decrypt does not support mauth, use default credential chain instead")
		}
		smClient, err := NewSecretManagerClientWithBearer(authConfig)
		if err != nil {
			return "", err
		}
		return smClient.GetSecretValue(ctx, secretName)
	}
	kmsClient, err := CreateKmsClient()
	if err != nil {
		return "", err
	}
	if command == CommandGetSecretsAndDecrypt {
		return GetSecretValueAndDecrypt(kmsClient, secretName)
	}
	return GetSecretValue(kmsClient, secretName)
}
