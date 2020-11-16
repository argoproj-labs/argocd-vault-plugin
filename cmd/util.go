package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	k8yaml "k8s.io/apimachinery/pkg/util/yaml"
)

func listYamlFiles(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".yaml" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return files, err
	}

	return files, nil
}

func readFilesAsManifests(paths []string) (result []map[string]interface{}, errs []error) {

	for _, path := range paths {
		manifest, err := manifestFromYAML(path)
		if err != nil {
			errs = append(errs, err)
		}
		result = append(result, manifest)
	}

	return result, errs
}

func manifestFromYAML(path string) (map[string]interface{}, error) {
	rawdata, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read YAML: %s from disk: %s", path, err)
	}

	decoder := k8yaml.NewYAMLOrJSONDecoder(bytes.NewReader(rawdata), 1000)
	var manifest map[string]interface{}
	err = decoder.Decode(&manifest)
	if err != nil {
		return nil, fmt.Errorf("could not read YAML: %s into a manifest: %s", path, err)
	}

	return manifest, nil
}
