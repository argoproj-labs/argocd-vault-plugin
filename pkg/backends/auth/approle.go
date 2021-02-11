package auth

import (
	"github.com/IBM/argocd-vault-plugin/pkg/utils"
	"github.com/hashicorp/vault/api"
)

// AppRole is a struct for working with Vault that uses AppRole
type AppRole struct {
	RoleID   string
	SecretID string
}

// Authenticate authenticates with Vault using App Role and returns a token
func (a *AppRole) Authenticate(vaultClient *api.Client) error {
	payload := map[string]interface{}{
		"role_id":   a.RoleID,
		"secret_id": a.SecretID,
	}
	data, err := vaultClient.Logical().Write("auth/approle/login", payload)
	if err != nil {
		return err
	}

	// If we cannot write the Vault token, we'll just have to login next time. Nothing showstopping.
	err = utils.SetToken(vaultClient, data.Auth.ClientToken)
	if err != nil {
		print(err)
	}

	return nil
}
