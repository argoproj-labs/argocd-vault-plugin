package cmd

import (
	"errors"
	"fmt"
	"os"

	kube "github.com/IBM/argocd-vault-plugin/pkg/kube"
	"github.com/IBM/argocd-vault-plugin/pkg/vault"
	"github.com/spf13/cobra"
)

// NewGenerateCommand initializes the generate command
func NewGenerateCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "generate <path>",
		Short: "Generate manifests from templates with Vault values",
		RunE: func(cmd *cobra.Command, args []string) error {
			vaultType := os.Getenv("VAULT_TYPE")
			if vaultType == "" {
				return errors.New("variable VAULT_TYPE was not set")
			}

			vaultClient, err := vault.InitVault(vaultType)
			if err != nil {
				return err
			}

			err = vaultClient.Login()
			if err != nil {
				return err
			}

			path := args[0]
			files, err := listYamlFiles(path)
			if len(files) < 1 {
				return fmt.Errorf("no YAML files were found in %s", path)
			}
			if err != nil {
				return err
			}

			manifests, errs := readFilesAsManifests(files)
			if len(errs) != 0 {

				// TODO: handle multiple errors nicely
				return fmt.Errorf("could not read YAML files: %s", errs)
			}

			for _, manifest := range manifests {

				resource, err := kube.CreateTemplate(manifest, vaultClient)
				if err != nil {
					return err
				}

				err = resource.Replace()
				if err != nil {
					return err
				}

				output, err := resource.ToYAML()
				if err != nil {
					return err
				}

				fmt.Printf("%s---\n", output)
			}

			return nil
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("<path> argument required to generate manifests")
			}
			return nil
		},
	}

	return command
}
