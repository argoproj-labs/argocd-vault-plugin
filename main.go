package main

import (
	"os"

	"github.com/IBM/argocd-vault-plugin/cmd"
)

func main() {
	if err := cmd.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
