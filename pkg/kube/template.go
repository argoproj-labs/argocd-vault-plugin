package kube

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/IBM/argocd-vault-plugin/pkg/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8yaml "k8s.io/apimachinery/pkg/util/yaml"
	yaml "sigs.k8s.io/yaml"
)

// A Resource is the basis for all Templates
type Resource struct {
	Kind              string
	TemplateData      unstructured.Unstructured // The template as read from YAML
	Backend           types.Backend
	replacementErrors []error                // Any errors encountered in performing replacements
	Data              map[string]interface{} // The data to replace with, from Vault
	Config            map[string]interface{}
}

// Template is the template for Kubernetes
type Template struct {
	Resource
}

// NewTemplate returns a *Template given the template's data, and a VaultType
func NewTemplate(template string, backend types.Backend) (*Template, error) {
	obj := &unstructured.Unstructured{}

	decoder := k8yaml.NewYAMLToJSONDecoder(bytes.NewReader([]byte(template)))
	err := decoder.Decode(&obj)
	if err != nil {
		return nil, fmt.Errorf("ToYAML: could not convert replaced template into %s: %s", obj.GetKind(), err)
	}

	var path string
	annotations := obj.GetAnnotations()
	if avpPath, ok := annotations["avp_path"]; ok {
		path = avpPath
	}

	var kvVersion string
	if kv, ok := annotations["kv_version"]; ok {
		kvVersion = kv
	}

	var avpIgnore bool
	if ignore, ok := annotations["avp_ignore"]; ok {
		avpIgnore, _ = strconv.ParseBool(ignore)
	}

	var data map[string]interface{}
	if path != "" {
		if !avpIgnore {
			data, err = backend.GetSecrets(path, kvVersion)
			if err != nil {
				return nil, err
			}
		}
	}

	return &Template{
		Resource{
			Kind:         obj.GetKind(),
			TemplateData: *obj,
			Backend:      backend,
			Data:         data,
			Config: map[string]interface{}{
				"kvVersion": kvVersion,
			},
		},
	}, nil
}

// Replace will replace the <placeholders> in the Template's data with values from Vault.
// It will return an aggregrate of any errors encountered during the replacements.
// For both non-Secret resources and Secrets with <placeholder>'s in `stringData`, the value in Vault is emitted as-is
// For Secret's with <placeholder>'s in `.data`, the value in Vault is emitted as base64
// For any hard-coded strings that aren't <placeholder>'s, the string is emitted as-is
func (t *Template) Replace() error {
	var replacerFunc func(string, string, Resource) (interface{}, []error)

	switch t.Kind {
	case "ConfigMap":
		replacerFunc = configReplacement
	default:
		replacerFunc = genericReplacement
	}

	replaceInner(&t.Resource, &t.TemplateData.Object, replacerFunc)
	if len(t.replacementErrors) != 0 {
		// TODO format multiple errors nicely
		return fmt.Errorf("Replace: could not replace all placeholders in Template: %s", t.replacementErrors)
	}
	return nil
}

// ToYAML seralizes the completed template into YAML to be consumed by Kubernetes
func (t *Template) ToYAML() (string, error) {
	res, err := yaml.Marshal(&t.TemplateData)
	if err != nil {
		return "", fmt.Errorf("ToYAML: could not export %s into YAML: %s", t.TemplateData.GetKind(), err)
	}
	return string(res), nil
}
