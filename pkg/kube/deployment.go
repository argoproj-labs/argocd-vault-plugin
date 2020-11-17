package kube

import (
	"fmt"

	"github.com/IBM/argocd-vault-plugin/pkg/vault"
	appsv1 "k8s.io/api/apps/v1"
	k8yaml "sigs.k8s.io/yaml"
)

// DeploymentTemplate is the template for Kubernetes Deployments
type DeploymentTemplate struct {
	Resource
}

// NewDeploymentTemplate returns a *DeploymentTemplate given the template's data, and a VaultType
func NewDeploymentTemplate(template map[string]interface{}, vault vault.VaultType) (*DeploymentTemplate, error) {

	data, err := vault.GetSecrets("/deployment")
	if err != nil {
		return nil, err
	}

	return &DeploymentTemplate{
		Resource{
			templateData: template,
			vaultData:    data,
		},
	}, nil
}

// Replace will replace the <placeholders> in the template's data with values from Vault.
// It will return an aggregrate of any errors encountered during the replacements
func (d *DeploymentTemplate) Replace() error {
	replaceInner(&d.Resource, &d.templateData, genericReplacement)
	if len(d.replacementErrors) != 0 {

		// TODO format multiple errors nicely
		return fmt.Errorf("Replace: could not replace all placeholders in DeploymentTemplate: %s", d.replacementErrors)
	}
	return nil
}

// ToYAML seralizes the completed template into YAML to be consumed by Kubernetes
func (d *DeploymentTemplate) ToYAML() (string, error) {
	kubeResource := appsv1.Deployment{}
	err := kubeResourceDecoder(&d.templateData).Decode(&kubeResource)
	if err != nil {
		return "", fmt.Errorf("ToYAML: could not convert replaced template into Deployment: %s", err)
	}
	res, err := k8yaml.Marshal(&kubeResource)
	if err != nil {
		return "", fmt.Errorf("ToYAML: could not export Deployment into YAML: %s", err)
	}
	return string(res), nil
}
