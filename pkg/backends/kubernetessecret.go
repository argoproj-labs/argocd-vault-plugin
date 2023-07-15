package backends

import (
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/kube"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"github.com/pkg/errors"
)

type kubeSecretsClient interface {
	ReadSecretData(string) (map[string][]byte, error)
}

// KubernetesSecret is a struct for working with a Kubernetes Secret backend
type KubernetesSecret struct {
	client kubeSecretsClient
}

// NewKubernetesSecret returns a new Kubernetes Secret backend.
func NewKubernetesSecret() *KubernetesSecret {
	return &KubernetesSecret{}
}

// Login initiates kubernetes client
func (k *KubernetesSecret) Login() error {
	localClient, err := kube.NewClient()
	if err != nil {
		return errors.Wrap(err, "Failed to perform login for kubernetes secret backend")
	}
	k.client = localClient
	return nil
}

// GetSecrets gets secrets from Kubernetes Secret and returns the formatted data
func (k *KubernetesSecret) GetSecrets(path string, version string, annotations map[string]string) (map[string]interface{}, error) {
	utils.VerboseToStdErr("K8s Secret getting secret: %s", path)
	data, err := k.client.ReadSecretData(path)
	if err != nil {
		return nil, err
	}

	out := make(map[string]interface{}, len(data))
	for k, v := range data {
		out[k] = string(v)
	}

	utils.VerboseToStdErr("K8s Secret get secret response: %v", out)
	return out, nil
}

// GetIndividualSecret will get the specific secret (placeholder) from the Kubernetes Secret backend
// Kubernetes Secrets can only be wholly read,
// So, we use GetSecrets and extract the specific placeholder we want
func (k *KubernetesSecret) GetIndividualSecret(path, secret, version string, annotations map[string]string) (interface{}, error) {
	utils.VerboseToStdErr("K8s Secret getting secret %s and key %s", path, secret)
	data, err := k.GetSecrets(path, version, annotations)
	if err != nil {
		return nil, err
	}
	return data[secret], nil
}
