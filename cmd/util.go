package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

func readFilesAsManifests(paths []string) (result []unstructured.Unstructured, errs []error) {

	for _, path := range paths {
		manifest, err := manifestFromYAML(path)
		if err != nil {
			errs = append(errs, err)
		}
		result = append(result, manifest...)
	}

	return result, errs
}

func manifestFromYAML(path string) ([]unstructured.Unstructured, error) {
	rawdata, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read YAML: %s from disk: %s", path, err)
	}

	decoder := k8yaml.NewYAMLToJSONDecoder(bytes.NewReader(rawdata))

	var manifests []unstructured.Unstructured
	for {
		nxtManifest := unstructured.Unstructured{}
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
