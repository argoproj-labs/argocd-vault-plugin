package kube

import (
	"fmt"

	"github.com/IBM/argocd-vault-plugin/pkg/vault"
)

// Template is created from YAML files for Kubernetes resources (manifests) that contain
// <placeholders>'s. Templates can be replaced by replacing the <placeholders>
// with values from Vault. They can be serialized back to YAML for usage by Kubernetes.
type Template interface {
	Replace() error
	ToYAML() (string, error)
}

// A Resource is the basis for all Templates
type Resource struct {
	templateData      map[string]interface{} // The template as read from YAML
	replacementErrors []error                // Any errors encountered in performing replacements
	vaultData         map[string]interface{} // The data to replace with, from Vault
}

// CreateTemplate will attempt to create the appropriate Template from a Kubernetes manifest.
// It will throw an error for unsupported manifest Kind's
func CreateTemplate(manifest map[string]interface{}, vault vault.VaultType) (Template, error) {
	switch manifest["kind"] {
	case "Deployment":
		{
			template, err := NewDeploymentTemplate(manifest, vault)
			if err != nil {
				return nil, err
			}
			return template, nil
		}
	case "Service":
		{
			template, err := NewServiceTemplate(manifest, vault)
			if err != nil {
				return nil, err
			}
			return template, nil
		}
	case "Secret":
		{
			template, err := NewSecretTemplate(manifest, vault)
			if err != nil {
				return nil, err
			}
			return template, nil
		}
	case "ConfigMap":
		{
			template, err := NewConfigMapTemplate(manifest, vault)
			if err != nil {
				return nil, err
			}
			return template, nil
		}
	default:
		{
			return nil, fmt.Errorf("unsupported kind: %s", manifest["kind"])
		}
	}
}
