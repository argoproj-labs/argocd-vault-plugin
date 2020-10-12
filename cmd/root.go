package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "argocd-vault",
	Short: "This is a plugin to replace <wildcards> with Vault secrets",
	Long:  "This is a plugin to replace <wildcards> with Vault secrets",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("The ArgoCD plugin is running")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
