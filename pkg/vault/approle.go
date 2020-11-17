package vault

import "github.com/IBM/argocd-vault-plugin/pkg/vault/client"

// AppRole is a struct for working with Vault that uses AppRole
type AppRole struct {
	RoleID   string
	SecretID string
	token    string
	*client.VaultClient
}

// Login authenticates with Vault using App Role and returns a token
func (a *AppRole) Login() error {
	payload := map[string]interface{}{
		"role_id":   a.RoleID,
		"secret_id": a.SecretID,
	}

	data, err := a.VaultClient.Write("auth/approle/login", payload)
	if err != nil {
		return err
	}

	a.token = data.Auth.ClientToken
	return nil
}

// GetSecrets gets secrets from vault and returns the formatted data
func (a *AppRole) GetSecrets(path string) (map[string]interface{}, error) {
	data, err := a.VaultClient.Read(path, a.token)
	if err != nil {
		return nil, err
	}

	return data, nil
}
