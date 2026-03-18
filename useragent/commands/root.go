package commands

import (
	"fmt"
	"os"

	"github.com/aliyun/alibabacloud-kms-cli/commands/openclaw"
	"github.com/aliyun/alibabacloud-kms-cli/commands/secret"
	"github.com/spf13/cobra"
)

func Execute() {

	var rootCmd = &cobra.Command{Use: "kmscli"}

	rootCmd.AddCommand(openclaw.OpenClawCmd)
	rootCmd.AddCommand(secret.SecretCmd)

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to execute cmd: %v\n", err)
		os.Exit(1)
	}
}
