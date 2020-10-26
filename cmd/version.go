package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewVersionCommand returns a new instance of the version command
func NewVersionCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "version",
		Short: "Print argocd-vault-plugin version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Version 1") // Just an example, should read from something
		},
	}

	return command
}
