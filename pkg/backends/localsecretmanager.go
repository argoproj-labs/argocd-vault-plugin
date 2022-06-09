package backends

import (
	"fmt"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type decryptFunc func(path string, filetype string) ([]byte, error)

// LocalSecretManager is a struct for working with local files
// Receives a function that knows how to decrypt the file, f.ex. using sops
type LocalSecretManager struct {
	Decrypt decryptFunc
}

// NewLocalSecretManagerBackend initializes a new local secret backend
func NewLocalSecretManagerBackend(decrypt decryptFunc) *LocalSecretManager {
	return &LocalSecretManager{
		Decrypt: decrypt,
	}
}

// Login does nothing as a "login" is handled by environment
func (a *LocalSecretManager) Login() error {
	return nil
}

// GetSecrets gets secrets using decrypt function and returns the formatted data
func (a *LocalSecretManager) GetSecrets(path string, version string, annotations map[string]string) (map[string]interface{}, error) {
	utils.VerboseToStdErr("Local secret manager getting secret %s at version %s", path, version)
	cleartext, err := a.Decrypt(path, "yaml")

	utils.VerboseToStdErr("Local secret manager get secret response: %v", cleartext)

	var dat map[string]interface{}

	if err != nil {
		return nil, fmt.Errorf("Could not find file %s - %s", path, err)
	}

	err = yaml.Unmarshal(cleartext, &dat)
	if err != nil {
		return nil, err
	}

	return dat, nil
}

// GetIndividualSecret will get the specific secret (placeholder) from the backend
// For local secrets, we only support placeholders replaced from the k/v pairs of a secret which cannot be individually addressed
// So, we use GetSecrets and extract the specific placeholder we want
func (a *LocalSecretManager) GetIndividualSecret(kvpath, secret, version string, annotations map[string]string) (interface{}, error) {
	data, err := a.GetSecrets(kvpath, version, annotations)
	if err != nil {
		return nil, err
	}
	return data[secret], nil
}
