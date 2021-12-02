package cmd

import (
	"fmt"

	"github.com/argoproj-labs/argocd-vault-plugin/version"
	"github.com/spf13/cobra"
)

// NewVersionCommand returns a new instance of the version command
func NewVersionCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "version",
		Short: "Print argocd-vault-plugin version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "argocd-vault-plugin %s (%s) BuildDate: %s\n", version.Version, version.CommitSHA, version.BuildDate)
		},
	}

	return command
}
