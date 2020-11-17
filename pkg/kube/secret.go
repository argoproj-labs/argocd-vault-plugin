package kube

import (
	"fmt"

	"github.com/IBM/argocd-vault-plugin/pkg/vault"
	corev1 "k8s.io/api/core/v1"
	k8yaml "sigs.k8s.io/yaml"
)

// SecretTemplate is the template for Kubernetes Secrets
type SecretTemplate struct {
	Resource
}

// NewSecretTemplate returns a *SecretTemplate given the template's data, and a VaultType
func NewSecretTemplate(template map[string]interface{}, prefix string, vault vault.VaultType) (*SecretTemplate, error) {
	data, err := vault.GetSecrets(prefix + "/secrets")
	if err != nil {
		return nil, err
	}

	return &SecretTemplate{
		Resource{
			templateData: template,
			vaultData:    data,
		},
	}, nil
}

// Replace will replace the <placeholders> in the template's data with values from Vault.
// It will return an aggregrate of any errors encountered during the replacements
func (d *SecretTemplate) Replace() error {

	// Replace metadata normally
	metadata, ok := d.templateData["metadata"].(map[string]interface{})
	if ok {
		replaceInner(&d.Resource, &metadata, genericReplacement)
		if len(d.replacementErrors) != 0 {

			// TODO format multiple errors nicely
			return fmt.Errorf("Replace: could not replace all placeholders in SecretTemplate metadata: %s", d.replacementErrors)
		}
	}

	// Replace the actual secrets with []byte's
	data, ok := d.templateData["data"].(map[string]interface{})
	if ok {
		replaceInner(&d.Resource, &data, secretReplacement)
		if len(d.replacementErrors) != 0 {

			// TODO format multiple errors nicely
			return fmt.Errorf("Replace: could not replace all placeholders in SecretTemplate data: %s", d.replacementErrors)
		}
	}

	return nil
}

// Ensures the replacements for the Secret data are in the right format
func secretReplacement(key, value string, vaultData map[string]interface{}) (_ interface{}, err []error) {
	res, err := genericReplacement(key, value, vaultData)

	// We have to return []byte for k8s secrets,
	// so we convert whatever is in Vault
	byteData := []byte(stringify(res))

	return byteData, err
}

// ToYAML seralizes the completed template into YAML to be consumed by Kubernetes
func (d *SecretTemplate) ToYAML() (string, error) {
	kubeResource := corev1.Secret{}
	err := kubeResourceDecoder(&d.templateData).Decode(&kubeResource)
	if err != nil {
		return "", fmt.Errorf("ToYAML: could not convert replaced template into Secret: %s", err)
	}
	res, err := k8yaml.Marshal(&kubeResource)
	if err != nil {
		return "", fmt.Errorf("ToYAML: could not export Secret into YAML: %s", err)
	}
	return string(res), nil
}
