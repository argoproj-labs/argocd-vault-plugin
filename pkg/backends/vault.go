package backends

import (
	"errors"
	"fmt"

	"github.com/IBM/argocd-vault-plugin/pkg/types"
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
func (v *Vault) GetSecrets(path, kvVersion string) (map[string]interface{}, error) {
	secret, err := v.VaultClient.Logical().Read(path)
	if err != nil {
		return nil, err
	}

	if secret == nil {
		return nil, fmt.Errorf("Could not find secrets at path %s", path)
	}

	if kvVersion != "" {
		v.KvVersion = kvVersion
	}

	if v.KvVersion == "2" {
		if _, ok := secret.Data["data"]; ok {
			return secret.Data["data"].(map[string]interface{}), nil
		}
		if len(secret.Data) == 0 {
			return nil, fmt.Errorf("The Vault path: %s is empty - did you forget to include /data/ in the Vault path for kv-v2?", path)
		}
		return nil, errors.New("Could not get data from Vault, check that kv-v2 is the correct engine")
	}

	if v.KvVersion == "1" {
		return secret.Data, nil
	}

	return nil, errors.New("Unsupported kvVersion specified")
}
