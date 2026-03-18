package secret

import (
	"fmt"
	"os"

	"github.com/aliyun/alibabacloud-kms-cli/kms"
	"github.com/spf13/cobra"
)

const commandGetSecrets = "getsecret"

var SecretCmd = &cobra.Command{
	Use:   "secret",
	Short: "get secrets from kms, return secret value by plain text format",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, fmt.Sprintf("Error: missing command arguments\nUsage: %s %s getsecret $secretName\n", os.Args[0], "secret"))
			os.Exit(1)
		}

		command := args[0]
		if command != commandGetSecrets {
			fmt.Fprintf(os.Stderr, fmt.Sprintf("Error: unknown command '%s', only 'getsecret' is supported\n", command))
			os.Exit(1)
		}

		secretName := args[1]
		if secretName == "" {
			fmt.Fprintf(os.Stderr, "Error: secret name cannot be empty\n")
			os.Exit(1)
		}

		value, err := runExecPlainMode(secretName)
		if err != nil {
			fmt.Fprintf(os.Stderr, fmt.Sprintf("Error: failed to get secret '%s': %v\n", secretName, err))
			os.Exit(1)
		}
		fmt.Print(string(value))
	},
}

func runExecPlainMode(secretName string) (value string, err error) {
	kmsClient, err := kms.CreateKmsClient()
	if err != nil {
		return "", err
	}
	secretValue, err := kms.GetSecretValue(kmsClient, secretName)
	if err != nil {
		return "", err
	}
	return secretValue, nil

}
