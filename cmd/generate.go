package cmd

import (
	"errors"
	"fmt"

	kube "github.com/IBM/argocd-vault-plugin/pkg/kube"
	"github.com/spf13/cobra"

	vault "github.com/IBM/argocd-vault-plugin/pkg/vault"
)

var (
	vaultAddress = "https://vserv-test.sos.ibm.com:8200"
)

// NewGenerateCommand initializes the generate command
func NewGenerateCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "generate <path>",
		Short: "Generate manifests from templates with Vault values",
		RunE: func(cmd *cobra.Command, args []string) error {
<<<<<<< HEAD

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

				resource, err := createTemplate(manifest)
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

			// check for env vars
			// get type of vault from env vars
			// perform appropriate authentication method if token doesn already exist
			// josh: read values from vault using token and then format into map string interface
			// josh: perform find and replace on yaml files
			// josh: output yaml files to standard out
			// path := args[0]
			// accessToken := args[1]
			//
			// githubClient := auth.GithubAuthMethod{
			// 	VaultAddress: vaultAddress,
			// }
			//
			// token, _ := githubClient.GetVaultToken(accessToken)
			//
			// results := getValues(path, token)
			// fmt.Print(results)
			// var thing vault.VaultType
			//
			// thing = &vault.Github{}
			// token, _ := thing.Login("token")
			// secrets, _ := thing.GetSecrets(token)
			// fmt.Print(secrets)
			return nil
		},
	}

	return command
}

<<<<<<< HEAD
func createTemplate(manifest map[string]interface{}) (kube.Template, error) {
	switch manifest["kind"] {
	case "Deployment":
		{
			return kube.NewDeploymentTemplate(manifest), nil
		}
	case "Secret":
		{
			return kube.NewSecretTemplate(manifest), nil
		}
	}
	return nil, fmt.Errorf("unsupported kind: %s", manifest["kind"])
}

// func getValues(path string, token string) map[string]interface{} {
// 	var httpClient = &http.Client{
// 		Timeout: 10 * time.Second,
// 	}
//
// 	client, err := vault.NewClient(&vault.Config{Address: vaultAddress, HttpClient: httpClient})
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	client.SetToken(token)
// 	data, err := client.Logical().Read("generic/user/jawernette/test-secret")
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	return data.Data
// }
