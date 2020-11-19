package kube

import (
	"fmt"

	"github.com/IBM/argocd-vault-plugin/pkg/vault"
	corev1 "k8s.io/api/core/v1"
	k8yaml "sigs.k8s.io/yaml"
)

// ConfigMapTemplate is the template for Kubernetes ConfigMaps
type ConfigMapTemplate struct {
	Resource
}

// NewConfigMapTemplate returns a *ConfigMapTemplate given the template's data, and a VaultType
func NewConfigMapTemplate(template map[string]interface{}, vault vault.VaultType) (*ConfigMapTemplate, error) {
	data, err := vault.GetSecrets("/config")
	if err != nil {
		return nil, err
	}

	return &ConfigMapTemplate{
		Resource{
			templateData: template,
			vaultData:    data,
		},
	}, nil
}

// Replace will replace the <placeholders> in the template's data with values from Vault.
// It will return an aggregrate of any errors encountered during the replacements
func (d *ConfigMapTemplate) Replace() error {

	// Assuming all other fields of a configMap (besides `data`) will have string values
	replaceInner(&d.Resource, &d.templateData, configMapReplacement)
	if len(d.replacementErrors) != 0 {

		// TODO format multiple errors nicely
		return fmt.Errorf("Replace: could not replace all placeholders in ConfigMapTemplate: %s", d.replacementErrors)
	}
	return nil
}

func configMapReplacement(key, value string, vaultData map[string]interface{}) (interface{}, []error) {
	res, err := genericReplacement(key, value, vaultData)
	if err != nil {
		return nil, err
	}
	// configMap data values must be strings
	return stringify(res), err
}

// ToYAML seralizes the completed template into YAML to be consumed by Kubernetes
func (d *ConfigMapTemplate) ToYAML() (string, error) {
	kubeResource := corev1.ConfigMap{}
	err := kubeResourceDecoder(&d.templateData).Decode(&kubeResource)
	if err != nil {
		return "", fmt.Errorf("ToYAML: could not convert replaced template into ConfigMap: %s", err)
	}
	res, err := k8yaml.Marshal(&kubeResource)
	if err != nil {
		return "", fmt.Errorf("ToYAML: could not export ConfigMap into YAML: %s", err)
	}
	return string(res), nil
}
