package openclaw

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/aliyun/alibabacloud-kms-cli/commands/common"
	"github.com/aliyun/alibabacloud-kms-cli/kms"
	"github.com/aliyun/alibabacloud-kms-cli/pkg/configfile"
	"github.com/aliyun/alibabacloud-kms-cli/pkg/mauth/config"
	"github.com/aliyun/alibabacloud-kms-cli/useragent"
	"github.com/spf13/cobra"
)

// kmscli openclaw getsecret

var (
	kmsEndpoint           string
	authMethod            string
	aapArn                string
	tokenPath             string
	clientKeyConfigPath   string
	clientKeyPasswordPath string
)

var OpenClawCmd = &cobra.Command{
	Use:   "openclaw",
	Short: "get openclaw secrets from kms, return json format",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			writeErrorResponse(globalErrorKey, fmt.Sprintf("Error: missing command arguments\nUsage: %s openclaw getsecret\n", os.Args[0]))
			os.Exit(1)
		}

		command := args[0]

		if command != commandGetSecrets && command != commandGetSecretsAndDecrypt {
			writeErrorResponse(globalErrorKey, fmt.Sprintf("Error: unknown command '%s', only 'getsecret' and 'getsecret-decrypt' are supported\n", command))
			os.Exit(1)
		}

		runExecProviderMode(command)
	},
}

func init() {
	common.RegisterKmsFlags(OpenClawCmd, &kmsEndpoint, &authMethod, &aapArn, &tokenPath, &clientKeyConfigPath, &clientKeyPasswordPath)
}

const (
	protocolVersion = 1

	globalErrorKey              = "global"
	commandGetSecrets           = "getsecret"
	commandGetSecretsAndDecrypt = "getsecret-decrypt"
	addUserAgent                = "openclaw"
)

// ExecProviderRequest Exec Provider 请求格式
type ExecProviderRequest struct {
	ProtocolVersion int      `json:"protocolVersion"`
	Provider        string   `json:"provider"`
	IDs             []string `json:"ids"`
}

// ExecProviderResponse Exec Provider 响应格式
type ExecProviderResponse struct {
	ProtocolVersion int                   `json:"protocolVersion"`
	Values          map[string]string     `json:"values"`
	Errors          map[string]*ErrorInfo `json:"errors,omitempty"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Message string `json:"message"`
}

// runExecProviderMode Exec Provider 模式（从 stdin 读取 JSON）
func runExecProviderMode(command string) {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		writeErrorResponse(globalErrorKey, fmt.Sprintf("Failed to read stdin: %v", err))
		os.Exit(1)
	}

	var req ExecProviderRequest
	if err := json.Unmarshal(input, &req); err != nil {
		writeErrorResponse(globalErrorKey, fmt.Sprintf("Failed to parse request: %v", err))
		os.Exit(1)
	}

	if req.ProtocolVersion != protocolVersion {
		writeErrorResponse(globalErrorKey, fmt.Sprintf("Unsupported protocol version: %d", req.ProtocolVersion))
		os.Exit(1)
	}
	useragent.RegisterUserAgent(addUserAgent)

	response := ExecProviderResponse{
		ProtocolVersion: protocolVersion,
		Values:          make(map[string]string),
		Errors:          make(map[string]*ErrorInfo),
	}

	authConfig := config.AuthConfig{
		AuthMethod:            authMethod,
		KmsEndpoint:           kmsEndpoint,
		AapArn:                aapArn,
		TokenPath:             tokenPath,
		ClientKeyConfigPath:   clientKeyConfigPath,
		ClientKeyPasswordPath: clientKeyPasswordPath,
	}

	cfgFile, err := configfile.Load()
	if err != nil {
		writeErrorResponse(globalErrorKey, fmt.Sprintf("Failed to load config file: %v", err))
		os.Exit(1)
	}
	common.MergeConfigFile(cfgFile, &authConfig)

	for _, id := range req.IDs {
		secretValue, err := kms.ResolveSecret(context.Background(), authConfig, command, id)
		if err != nil {
			response.Errors[id] = &ErrorInfo{
				Message: fmt.Sprintf("Failed to get secret: %v", err),
			}
		} else {
			response.Values[id] = secretValue
		}
	}
	output, err := json.Marshal(response)
	if err != nil {
		writeErrorResponse(globalErrorKey, fmt.Sprintf("Failed to serialize response: %v", err))
		os.Exit(1)
	}
	fmt.Print(string(output))
}

func writeErrorResponse(id string, message string) {
	response := ExecProviderResponse{
		ProtocolVersion: protocolVersion,
		Values:          make(map[string]string),
		Errors: map[string]*ErrorInfo{
			id: {Message: message},
		},
	}
	output, err := json.Marshal(response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal response error %v\n", err)
		os.Exit(1)
	}
	fmt.Print(string(output))
}
