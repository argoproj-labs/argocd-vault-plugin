package kube

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/IBM/argocd-vault-plugin/pkg/vault"
	appsv1 "k8s.io/api/apps/v1"
	k8yamldecoder "k8s.io/apimachinery/pkg/util/yaml"
	k8yaml "sigs.k8s.io/yaml"
)

// DeploymentTemplate is the template for Kubernetes Deployments
type DeploymentTemplate struct {
	Resource
}

// NewDeploymentTemplate returns a *DeploymentTemplate given the template's data, and a VaultType
func NewDeploymentTemplate(template map[string]interface{}, vault vault.VaultType) (*DeploymentTemplate, error) {
	path := os.Getenv("VAULT_PATH_PREFIX")
	data, err := vault.GetSecrets(path)
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
	jsondata, _ := json.Marshal(d.templateData)
	decoder := k8yamldecoder.NewYAMLOrJSONDecoder(bytes.NewReader(jsondata), 1000)
	kubeResource := appsv1.Deployment{}
	err := decoder.Decode(&kubeResource)
	if err != nil {
		return "", fmt.Errorf("ToYAML: could not convert replaced template into Deployment: %s", err)
	}
	res, err := k8yaml.Marshal(&kubeResource)
	if err != nil {
		return "", fmt.Errorf("ToYAML: could not export Deployment into YAML: %s", err)
	}
	return string(res), nil
}
