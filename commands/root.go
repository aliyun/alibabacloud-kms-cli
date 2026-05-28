package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/aliyun/alibabacloud-kms-cli/commands/openclaw"
	"github.com/aliyun/alibabacloud-kms-cli/commands/secret"
	"github.com/spf13/cobra"
)

const NameEnv = "KMSCLI_NAME"

func Execute() {
	name := os.Getenv(NameEnv)
	rootCmd := &cobra.Command{Use: "kmscli"}
	if name != "" {
		old := rootCmd.UsageFunc()
		rootCmd.SetUsageFunc(func(c *cobra.Command) error {
			var b strings.Builder
			oldOut := c.OutOrStdout()
			c.SetOut(&b)
			defer c.SetOut(oldOut)
			if err := old(c); err != nil {
				return err
			}
			path := name
			if c.Parent() != nil {
				path += " " + c.Name()
			}
			fmt.Fprint(oldOut, strings.ReplaceAll(b.String(), c.CommandPath(), path))
			return nil
		})
	}

	rootCmd.AddCommand(openclaw.OpenClawCmd)
	rootCmd.AddCommand(secret.SecretCmd)

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to execute cmd: %v\n", err)
		os.Exit(1)
	}
}
