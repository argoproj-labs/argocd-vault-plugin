package cmd

import (
	"fmt"

	kube "github.com/IBM/argocd-vault-plugin/pkg/kube"
	"github.com/IBM/argocd-vault-plugin/pkg/vault"
	"github.com/spf13/cobra"
)

// NewGenerateCommand initializes the generate command
func NewGenerateCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "generate <path>",
		Short: "Generate manifests from templates with Vault values",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("<path> argument required to generate manifests")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
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

			config, err := vault.NewConfig()
			if err != nil {
				return err
			}

			vaultClient := config.Type
			err = vaultClient.Login()
			if err != nil {
				return err
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
	}

	return command
}
