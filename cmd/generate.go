package cmd

import (
	"fmt"

	"github.com/IBM/argocd-vault-plugin/pkg/kube"
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

				// obj := &unstructured.Unstructured{}
				// err := kube.KubeResourceDecoder(&manifest).Decode(&obj)
				// if err != nil {
				// 	return fmt.Errorf("ToYAML: could not convert replaced template into %s: %s", obj.GetKind(), err)
				// }
				//
				// path := fmt.Sprintf("%s/%s", config.PathPrefix, obj.GetKind())
				//
				// annotations := obj.GetAnnotations()
				// if avpPath, ok := annotations["avp_path"]; ok {
				// 	path = avpPath
				// }
				//
				// vaultData, err := vaultClient.GetSecrets(path)
				// if err != nil {
				// 	return err
				// }
				//
				// template := &kube.Template{
				// 	Resource: kube.Resource{
				// 		Kind:         obj.GetKind(),
				// 		TemplateData: manifest,
				// 		VaultData:    vaultData,
				// 	},
				// }
				template, err := kube.NewTemplate(manifest, vaultClient, config.PathPrefix)
				if err != nil {
					return err
				}

				err = template.Replace()
				if err != nil {
					return err
				}

				output, err := template.ToYAML()
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
