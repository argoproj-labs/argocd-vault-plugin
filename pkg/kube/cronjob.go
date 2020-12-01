package kube

import (
	"fmt"

	"github.com/IBM/argocd-vault-plugin/pkg/vault"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	k8yaml "sigs.k8s.io/yaml"
)

// CronJobTemplate is the template for Kubernetes CronJobs
type CronJobTemplate struct {
	Resource
}

// NewCronJobTemplate returns a *CronJobTemplate given the template's data, and a VaultType
func NewCronJobTemplate(template map[string]interface{}, vault vault.VaultType) (*CronJobTemplate, error) {

	data, err := vault.GetSecrets("/deployment")
	if err != nil {
		return nil, err
	}

	return &CronJobTemplate{
		Resource{
			templateData: template,
			vaultData:    data,
		},
	}, nil
}

// Replace will replace the <placeholders> in the template's data with values from Vault.
// It will return an aggregrate of any errors encountered during the replacements
func (d *CronJobTemplate) Replace() error {
	replaceInner(&d.Resource, &d.templateData, genericReplacement)
	if len(d.replacementErrors) != 0 {

		// TODO format multiple errors nicely
		return fmt.Errorf("Replace: could not replace all placeholders in CronJobTemplate: %s", d.replacementErrors)
	}
	return nil
}

// ToYAML seralizes the completed template into YAML to be consumed by Kubernetes
func (d *CronJobTemplate) ToYAML() (string, error) {
	kubeResource := batchv1beta1.CronJob{}
	err := kubeResourceDecoder(&d.templateData).Decode(&kubeResource)
	if err != nil {
		return "", fmt.Errorf("ToYAML: could not convert replaced template into CronJob: %s", err)
	}
	res, err := k8yaml.Marshal(&kubeResource)
	if err != nil {
		return "", fmt.Errorf("ToYAML: could not export CronJob into YAML: %s", err)
	}
	return string(res), nil
}
