package secret

import (
	"context"
	"fmt"
	"os"

	"github.com/aliyun/alibabacloud-kms-cli/commands/common"
	"github.com/aliyun/alibabacloud-kms-cli/kms"
	"github.com/aliyun/alibabacloud-kms-cli/pkg/configfile"
	"github.com/aliyun/alibabacloud-kms-cli/pkg/mauth/config"
	"github.com/spf13/cobra"
)

const commandGetSecrets = "getsecret"
const commandGetSecretsAndDecrypt = "getsecret-decrypt"

var (
	kmsEndpoint           string
	authMethod            string
	aapArn                string
	tokenPath             string
	clientKeyConfigPath   string
	clientKeyPasswordPath string
)

var SecretCmd = &cobra.Command{
	Use:   "secret",
	Short: "get secrets from kms, return secret value by plain text format",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Error: missing command arguments\nUsage: %s %s getsecret $secretName\n", os.Args[0], "secret")
			os.Exit(1)
		}

		command := args[0]
		if command != commandGetSecrets && command != commandGetSecretsAndDecrypt {
			fmt.Fprintf(os.Stderr, "Error: unknown command '%s', only 'getsecret' and 'getsecret-decrypt' are supported\n", command)
			os.Exit(1)
		}

		secretName := args[1]
		if secretName == "" {
			fmt.Fprintf(os.Stderr, "Error: secret name cannot be empty\n")
			os.Exit(1)
		}

		value, err := runExecPlainMode(secretName, command)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to get secret '%s': %v\n", secretName, err)
			os.Exit(1)
		}
		fmt.Print(value)
	},
}

func init() {
	common.RegisterKmsFlags(SecretCmd, &kmsEndpoint, &authMethod, &aapArn, &tokenPath, &clientKeyConfigPath, &clientKeyPasswordPath)
}

func runExecPlainMode(secretName, command string) (value string, err error) {
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
		return "", fmt.Errorf("failed to load config file: %w", err)
	}
	common.MergeConfigFile(cfgFile, &authConfig)

	return kms.ResolveSecret(context.Background(), authConfig, command, secretName)
}
