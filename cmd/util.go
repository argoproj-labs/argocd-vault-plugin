package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	k8yaml "k8s.io/apimachinery/pkg/util/yaml"
)

func listYamlFiles(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
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
		result = append(result, manifest...)
	}

	return result, errs
}

func manifestFromYAML(path string) ([]map[string]interface{}, error) {
	rawdata, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read YAML: %s from disk: %s", path, err)
	}
	decoder := k8yaml.NewYAMLToJSONDecoder(bytes.NewReader(rawdata))
	var manifests []map[string]interface{}

	for {
		nxtManifest := make(map[string]interface{})
		err := decoder.Decode(&nxtManifest)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("could not read YAML: %s into a manifest: %s", path, err)
		}
		manifests = append(manifests, nxtManifest)
	}

	return manifests, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
