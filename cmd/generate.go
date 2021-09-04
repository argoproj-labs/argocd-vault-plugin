package cmd

import (
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/IBM/argocd-vault-plugin/pkg/config"
	"github.com/IBM/argocd-vault-plugin/pkg/kube"
	"github.com/IBM/argocd-vault-plugin/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewGenerateCommand initializes the generate command
func NewGenerateCommand() *cobra.Command {
	const StdIn = "-"
	var configPath, secretName string
	var useServiceAccountNamespace bool

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
			var manifests []unstructured.Unstructured
			var err error

			path := args[0]
			if path == StdIn {
				manifests, err = readManifestData(cmd.InOrStdin())
				if err != nil {
					return err
				}
			} else {
				files, err := listFiles(path)
				if len(files) < 1 {
					return fmt.Errorf("no YAML or JSON files were found in %s", path)
				}
				if err != nil {
					return err
				}

				var errs []error
				manifests, errs = readFilesAsManifests(files)
				if len(errs) != 0 {
					// TODO: handle multiple errors nicely
					return fmt.Errorf("could not read YAML/JSON files: %s", errs)
				}
			}

			v := viper.New()
			cmdConfig, err := config.New(v, &config.Options{
				SecretName:                 secretName,
				UseServiceAccountNamespace: useServiceAccountNamespace,
				ConfigPath:                 configPath,
			})
			if err != nil {
				return err
			}

			err = cmdConfig.Backend.Login()
			if err != nil {
				return err
			}

			for _, manifest := range manifests {

				if len(manifest.Object) == 0 {
					continue
				}

				template, err := kube.NewTemplate(manifest, cmdConfig.Backend)
				if err != nil {
					return err
				}

				annotations := manifest.GetAnnotations()
				avpIgnore, _ := strconv.ParseBool(annotations[types.AVPIgnoreAnnotation])
				if !avpIgnore {
					err = template.Replace()
					if err != nil {
						return err
					}
				}

				output, err := template.ToYAML()
				if err != nil {
					return err
				}

				fmt.Fprintf(cmd.OutOrStdout(), "%s---\n", output)
			}

			return nil
		},
	}

	command.Flags().StringVarP(&configPath, "config-path", "c", "", "path to a file containing Vault configuration (YAML, JSON, envfile) to use")
	command.Flags().StringVarP(&secretName, "secret-name", "s", "", "name of a Kubernetes Secret containing Vault configuration data, use 'namespace/secret-name' to specify the namespace (Only available when used in ArgoCD)")
	command.Flags().BoolVarP(&useServiceAccountNamespace, "service-account-namespace", "a", false, "when used with -s <secret name>, grab the secret from the namespace of the service account")
	return command
}
