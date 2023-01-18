package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/config"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/kube"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/types"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewGenerateCommand initializes the generate command
func NewGenerateCommand() *cobra.Command {
	const StdIn = "-"
	var configPath, secretName string
	var verboseOutput bool

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
					errMessages := make([]string, len(errs))
					for idx, err := range errs {
						errMessages[idx] = err.Error()
					}
					return fmt.Errorf("could not read YAML/JSON files:\n%s", strings.Join(errMessages, "\n"))
				}
			}

			v := viper.New()
			viper.Set("verboseOutput", verboseOutput)
			cmdConfig, err := config.New(v, &config.Options{
				SecretName: secretName,
				ConfigPath: configPath,
			})
			if err != nil {
				return err
			}

			err = cmdConfig.Backend.Login()
			if err != nil {
				return err
			}

			for _, manifest := range manifests {
				var pathValidation *regexp.Regexp
				if rexp := v.GetString(types.EnvPathValidation); rexp != "" {
					pathValidation, err = regexp.Compile(rexp)
					if err != nil {
						return fmt.Errorf("%s is not a valid regular expression: %s", rexp, err)
					}
				}

				template, err := kube.NewTemplate(manifest, cmdConfig.Backend, pathValidation)
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
				} else {
					utils.VerboseToStdErr("skipping %s.%s because %s annotation is present", manifest.GetNamespace(), manifest.GetName(), types.AVPIgnoreAnnotation)
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
	command.Flags().StringVarP(&secretName, "secret-name", "s", "", "name of a Kubernetes Secret in the argocd namespace containing Vault configuration data in the argocd namespace of your ArgoCD host (Only available when used in ArgoCD). The namespace can be overridden by using the format <namespace>:<name>")
	command.Flags().BoolVar(&verboseOutput, "verbose-sensitive-output", false, "enable verbose mode for detailed info to help with debugging. Includes sensitive data (credentials), logged to stderr")
	return command
}
