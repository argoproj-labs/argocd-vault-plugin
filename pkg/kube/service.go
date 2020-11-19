package kube

import (
	"fmt"

	"github.com/IBM/argocd-vault-plugin/pkg/vault"
	corev1 "k8s.io/api/core/v1"
	k8yaml "sigs.k8s.io/yaml"
)

// ServiceTemplate is the template for Kubernetes Services
type ServiceTemplate struct {
	Resource
}

// NewServiceTemplate returns a *ServiceTemplate given the template's data, and a VaultType
func NewServiceTemplate(template map[string]interface{}, vault vault.VaultType) (*ServiceTemplate, error) {
	data, err := vault.GetSecrets("/service")
	if err != nil {
		return nil, err
	}

	return &ServiceTemplate{
		Resource{
			templateData: template,
			vaultData:    data,
		},
	}, nil
}

// Replace will replace the <placeholders> in the template's data with values from Vault.
// It will return an aggregrate of any errors encountered during the replacements
func (d *ServiceTemplate) Replace() error {
	replaceInner(&d.Resource, &d.templateData, genericReplacement)
	if len(d.replacementErrors) != 0 {

		// TODO format multiple errors nicely
		return fmt.Errorf("Replace: could not replace all placeholders in ServiceTemplate: %s", d.replacementErrors)
	}
	return nil
}

// ToYAML seralizes the completed template into YAML to be consumed by Kubernetes
func (d *ServiceTemplate) ToYAML() (string, error) {
	kubeResource := corev1.Service{}
	err := kubeResourceDecoder(&d.templateData).Decode(&kubeResource)
	if err != nil {
		return "", fmt.Errorf("ToYAML: could not convert replaced template into Service: %s", err)
	}
	res, err := k8yaml.Marshal(&kubeResource)
	if err != nil {
		return "", fmt.Errorf("ToYAML: could not export Service into YAML: %s", err)
	}
	return string(res), nil
}
