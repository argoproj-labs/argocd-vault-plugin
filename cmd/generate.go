package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewGenerateCommand Initializes the generate command
func NewGenerateCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "generate",
		Short: "Generate YAML files from template with Vault values",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprint(cmd.OutOrStdout(), "Generate YAML")
			return nil
		},
	}

	return command
}
