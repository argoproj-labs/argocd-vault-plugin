package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

func readFilesAsManifests(paths []string) (result []string, errs []error) {

	for _, path := range paths {
		manifest, err := manifestFromYAML(path)
		if err != nil {
			errs = append(errs, err)
		}
		result = append(result, manifest...)
	}

	return result, errs
}

func manifestFromYAML(path string) ([]string, error) {
	rawdata, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read YAML: %s from disk: %s", path, err)
	}

	manifests := strings.Split(string(rawdata), "\n---")

	res := []string{}
	for _, doc := range manifests {
		content := strings.TrimSpace(doc)
		// Ignore empty docs
		if content != "" {
			res = append(res, content+"\n")
		}
	}

	return res, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
