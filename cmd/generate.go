package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	k8yaml "k8s.io/apimachinery/pkg/util/yaml"

	kube "github.com/IBM/argocd-vault-plugin/pkg/kube"
	"github.com/spf13/cobra"
)

// NewGenerateCommand Initializes the generate command
func NewGenerateCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "generate <path>",
		Short: "Generate manifests from templates with Vault values",
		RunE: func(cmd *cobra.Command, args []string) error {

			path := args[0]
			files := listYamlFiles(path)
			if len(files) < 1 {
				return fmt.Errorf("No YAML files were found in %s", path)
			}
			manifests := readFilesAsManifests(files)
			var resource kube.Template

			for _, manifest := range manifests {
				switch manifest["kind"] {
				case "Deployment":
					{
						resource = kube.NewDeploymentTemplate(manifest)
					}
				case "Secret":
					{
						resource = kube.NewSecretTemplate(manifest)
					}
				}

				err := resource.Replace()
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

func listYamlFiles(root string) []string {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".yaml" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	return files
}

func readFilesAsManifests(paths []string) []map[string]interface{} {
	var result []map[string]interface{}

	for _, path := range paths {
		result = append(result, manifestFromYaml(path))
	}

	return result
}

func manifestFromYaml(path string) map[string]interface{} {
	// Read as byte string
	rawdata, err := ioutil.ReadFile(path)
	if err != nil {

	}

	decoder := k8yaml.NewYAMLOrJSONDecoder(bytes.NewReader(rawdata), 1000)
	var manifest map[string]interface{}
	_ = decoder.Decode(&manifest)

	return manifest
}
