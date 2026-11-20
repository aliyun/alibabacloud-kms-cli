package kms

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/aliyun/alibabacloud-kms-cli/pkg/constants"
)

const (
	metadataURL        = "http://100.100.100.200/latest/meta-data/region-id"
	kmsVpcEndpoint     = "kms-vpc.%s.aliyuncs.com"
	kmsPublicEndpoint  = "kms.%s.aliyuncs.com"
	httpTimeout        = 5 * time.Second
	EndpointTypePublic = "Public"
	EndpointTypeVpc    = "Vpc"
)

// ResolveKmsEndpoint 解析 KMS Endpoint。
// 如果传入的 endpoint 不为空则直接返回；否则根据 REGION_ID / ALIBABA_CLOUD_REGION_ID 环境变量或 ECS metadata 自动获取 regionId 并拼接。
func ResolveKmsEndpoint(kmsEndpoint string) (string, error) {
	if kmsEndpoint != "" {
		return kmsEndpoint, nil
	}
	regionId, err := getRegionId()
	if err != nil {
		return "", fmt.Errorf("get endpoint and region id err: %v", err)
	}
	endpointType := getEndpointType()
	if endpointType == EndpointTypePublic {
		return fmt.Sprintf(kmsPublicEndpoint, regionId), nil
	}
	return fmt.Sprintf(kmsVpcEndpoint, regionId), nil
}

func getEndpointType() string {
	if os.Getenv(constants.EnvEndpointType) != "" {
		return os.Getenv(constants.EnvEndpointType)
	}
	return EndpointTypeVpc
}

func getRegionId() (string, error) {
	// get from env
	if regionId := os.Getenv(constants.EnvRegionID); regionId != "" {
		return regionId, nil
	}
	if regionId := os.Getenv(constants.EnvAlibabaCloudRegionID); regionId != "" {
		return regionId, nil
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
