package kube

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/types"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yaml "sigs.k8s.io/yaml"
)

// A Resource is the basis for all Templates
type Resource struct {
	Kind              string
	TemplateData      map[string]interface{} // The template as read from YAML
	Backend           types.Backend
	replacementErrors []error                // Any errors encountered in performing replacements
	Data              map[string]interface{} // The data to replace with, from Vault
	Annotations       map[string]string
	PathValidation    *regexp.Regexp
}

// Template is the template for Kubernetes
type Template struct {
	Resource
}

// NewTemplate returns a *Template given the template's data, and a VaultType
func NewTemplate(template unstructured.Unstructured, backend types.Backend, pathValidation *regexp.Regexp) (*Template, error) {
	annotations := template.GetAnnotations()
	path := annotations[types.AVPPathAnnotation]
	version := annotations[types.AVPSecretVersionAnnotation]

	var err error
	var data map[string]interface{}
	if path != "" {
		if pathValidation != nil && !pathValidation.MatchString(path) {
			return nil, fmt.Errorf("the path %s is disallowed by %s restriction", path, types.EnvPathValidation)
		}
		data, err = backend.GetSecrets(path, version, annotations)
		if err != nil {
			return nil, err
		}

		utils.VerboseToStdErr("calling GetSecrets to get all secrets from backend because %s is set to %s", types.AVPPathAnnotation, path)
	}

	return &Template{
		Resource{
			Kind:           template.GetKind(),
			TemplateData:   template.Object,
			Backend:        backend,
			Data:           data,
			Annotations:    annotations,
			PathValidation: pathValidation,
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
	case "Secret":
		replacerFunc = secretReplacement
	default:
		replacerFunc = genericReplacement
	}

	replaceInner(&t.Resource, &t.TemplateData, replacerFunc)
	if len(t.replacementErrors) != 0 {
		errMessages := make([]string, len(t.replacementErrors))
		for idx, err := range t.replacementErrors {
			errMessages[idx] = err.Error()
		}
		return fmt.Errorf("Replace: could not replace all placeholders in Template:\n%s", strings.Join(errMessages, "\n"))
	}
	return nil
}

// ToYAML serializes the completed template into YAML to be consumed by Kubernetes
func (t *Template) ToYAML() (string, error) {
	res, err := yaml.Marshal(&t.TemplateData)
	if err != nil {
		return "", fmt.Errorf("ToYAML: could not export %s into YAML: %s", t.Kind, err)
	}
	return string(res), nil
}
