package kms

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/aliyun/alibabacloud-kms-cli/useragent"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	kms20160120 "github.com/alibabacloud-go/kms-20160120/v3/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/credentials-go/credentials"
)

const (
	metadataURL        = "http://100.100.100.200/latest/meta-data/region-id"
	kmsVpcEndpoint     = "kms-vpc.%s.aliyuncs.com"
	kmsPublicEndpoint  = "kms.%s.aliyuncs.com"
	httpTimeout        = 5 * time.Second
	EndpointTypePublic = "Public"
	EndpointTypeVpc    = "Vpc"
)

const (
	InstanceGatewayDomainSuffix = "cryptoservice.kms.aliyuncs.com"
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
	regionId, err := getRegionId()
	if err != nil {
		return result, fmt.Errorf("get endpoint and region id err: %v", err)
	}

	endpointType := getEndpointType()

	if endpointType == EndpointTypePublic {
		config.Endpoint = tea.String(fmt.Sprintf(kmsPublicEndpoint, regionId))
	} else {
		config.Endpoint = tea.String(fmt.Sprintf(kmsVpcEndpoint, regionId))
	}

	return kms20160120.NewClient(config)
}

func getEndpointType() string {
	if os.Getenv("ENDPOINT_TYPE") != "" {
		return os.Getenv("ENDPOINT_TYPE")
	}
	return EndpointTypeVpc
}

func getRegionId() (string, error) {

	// get from env
	if os.Getenv("REGION_ID") != "" {
		return os.Getenv("REGION_ID"), nil
	}

	// get from metadata service
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(metadataURL)
	if err != nil {
		return "", fmt.Errorf("get region id from meta server error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get region id from meta server status invalid: %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
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
