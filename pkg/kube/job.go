package kube

import (
	"fmt"

	"github.com/IBM/argocd-vault-plugin/pkg/vault"
	batchv1 "k8s.io/api/batch/v1"
	k8yaml "sigs.k8s.io/yaml"
)

// JobTemplate is the template for Kubernetes Jobs
type JobTemplate struct {
	Resource
}

// NewJobTemplate returns a *JobTemplate given the template's data, and a VaultType
func NewJobTemplate(template map[string]interface{}, vault vault.VaultType) (*JobTemplate, error) {

	data, err := vault.GetSecrets("/deployment")
	if err != nil {
		return nil, err
	}

	return &JobTemplate{
		Resource{
			templateData: template,
			vaultData:    data,
		},
	}, nil
}

// Replace will replace the <placeholders> in the template's data with values from Vault.
// It will return an aggregrate of any errors encountered during the replacements
func (d *JobTemplate) Replace() error {
	replaceInner(&d.Resource, &d.templateData, genericReplacement)
	if len(d.replacementErrors) != 0 {

		// TODO format multiple errors nicely
		return fmt.Errorf("Replace: could not replace all placeholders in JobTemplate: %s", d.replacementErrors)
	}
	return nil
}

// ToYAML seralizes the completed template into YAML to be consumed by Kubernetes
func (d *JobTemplate) ToYAML() (string, error) {
	kubeResource := batchv1.Job{}
	err := kubeResourceDecoder(&d.templateData).Decode(&kubeResource)
	if err != nil {
		return "", fmt.Errorf("ToYAML: could not convert replaced template into Job: %s", err)
	}
	res, err := k8yaml.Marshal(&kubeResource)
	if err != nil {
		return "", fmt.Errorf("ToYAML: could not export Job into YAML: %s", err)
	}
	return string(res), nil
}
