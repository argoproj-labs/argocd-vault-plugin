package kube

import (
	"fmt"

	"github.com/IBM/argocd-vault-plugin/pkg/vault"
	networkingv1 "k8s.io/api/networking/v1"
	k8yaml "sigs.k8s.io/yaml"
)

// IngressTemplate is the template for Kubernetes Ingresses
type IngressTemplate struct {
	Resource
}

// NewIngressTemplate returns a *IngressTemplate given the template's data, and a VaultType
func NewIngressTemplate(template map[string]interface{}, vault vault.VaultType) (*IngressTemplate, error) {

	data, err := vault.GetSecrets("/ingress")
	if err != nil {
		return nil, err
	}

	return &IngressTemplate{
		Resource{
			templateData: template,
			vaultData:    data,
		},
	}, nil
}

// Replace will replace the <placeholders> in the template's data with values from Vault.
// It will return an aggregrate of any errors encountered during the replacements
func (d *IngressTemplate) Replace() error {
	replaceInner(&d.Resource, &d.templateData, genericReplacement)
	if len(d.replacementErrors) != 0 {

		// TODO format multiple errors nicely
		return fmt.Errorf("Replace: could not replace all placeholders in IngressTemplate: %s", d.replacementErrors)
	}
	return nil
}

// ToYAML seralizes the completed template into YAML to be consumed by Kubernetes
func (d *IngressTemplate) ToYAML() (string, error) {
	kubeResource := networkingv1.Ingress{}
	err := kubeResourceDecoder(&d.templateData).Decode(&kubeResource)
	if err != nil {
		return "", fmt.Errorf("ToYAML: could not convert replaced template into Ingress: %s", err)
	}
	res, err := k8yaml.Marshal(&kubeResource)
	if err != nil {
		return "", fmt.Errorf("ToYAML: could not export Ingress into YAML: %s", err)
	}
	return string(res), nil
}
