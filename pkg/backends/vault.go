package backends

import (
	"errors"
	"fmt"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/types"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"github.com/hashicorp/vault/api"
)

// Vault is a struct for working with a Vault backend
type Vault struct {
	types.AuthType
	VaultClient *api.Client
	KvVersion   string
}

// NewVaultBackend initializes a new Vault Backend
func NewVaultBackend(auth types.AuthType, client *api.Client, kv string) *Vault {
	vault := &Vault{
		KvVersion:   kv,
		AuthType:    auth,
		VaultClient: client,
	}

	return vault
}

// Login authenticates with the auth type provided
func (v *Vault) Login() error {
	err := v.AuthType.Authenticate(v.VaultClient)
	if err != nil {
		return err
	}
	return nil
}

// GetSecrets gets secrets from vault and returns the formatted data
func (v *Vault) GetSecrets(path string, version string, annotations map[string]string) (map[string]interface{}, error) {
	var secret *api.Secret
	var err error

	var kvVersion = v.KvVersion
	if kv, ok := annotations[types.VaultKVVersionAnnotation]; ok {
		kvVersion = kv
	}

	// Vault KV-V1 doesn't support versioning so we only honor `version` if KV-V2 is used
	if version != "" && kvVersion == "2" {
		utils.VerboseToStdErr("Hashicorp Vault getting kv pairs from KV-V2 path %s at version %s", path, version)
		secret, err = v.VaultClient.Logical().ReadWithData(path, map[string][]string{
			"version": {version},
		})
	} else {
		utils.VerboseToStdErr("Hashicorp Vault getting kv pairs from KV-V1 path %s", path)
		secret, err = v.VaultClient.Logical().Read(path)
	}

	if err != nil {
		return nil, err
	}

	utils.VerboseToStdErr("Hashicorp Vault get kv pairs response: %v", secret)

	if secret == nil {
		// Do not mention `version` in error message when it's not honored (KV-V1)
		if version == "" || kvVersion == "1" {
			return nil, fmt.Errorf("Could not find secrets at path %s", path)
		}
		return nil, fmt.Errorf("Could not find secrets at path %s with version %s", path, version)
	}

	if kvVersion == "2" {
		if _, ok := secret.Data["data"]; ok {
			if secret.Data["data"] != nil {
				return secret.Data["data"].(map[string]interface{}), nil
			}
			return nil, fmt.Errorf("The secret version %s for Vault path %s is nil - is this version of the secret deleted?", version, path)
		}
		if len(secret.Data) == 0 {
			return nil, fmt.Errorf("The Vault path: %s is empty - did you forget to include /data/ in the Vault path for kv-v2?", path)
		}
		return nil, errors.New("Could not get data from Vault, check that kv-v2 is the correct engine")
	}

	if kvVersion == "1" {
		return secret.Data, nil
	}

	return nil, errors.New("Unsupported kvVersion specified")
}

// GetIndividualSecret will get the specific secret (placeholder) from the SM backend
// For Vault, we only support placeholders replaced from the k/v pairs of a secret which cannot be individually addressed
// So, we use GetSecrets and extract the specific placeholder we want
func (v *Vault) GetIndividualSecret(kvpath, secret, version string, annotations map[string]string) (interface{}, error) {
	data, err := v.GetSecrets(kvpath, version, annotations)
	if err != nil {
		return nil, err
	}
	return data[secret], nil
}
