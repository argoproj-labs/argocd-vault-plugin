package vault

import (
	"github.com/IBM/argocd-vault-plugin/pkg/utils"
	"github.com/hashicorp/vault/api"
)

// Just a plain vault token
type TokenAuth struct {
	AccessToken string
}

// We just want to pass-through the vault token here
func NewTokenAuth(token string) *TokenAuth {
	tokenAuth := &TokenAuth{
		AccessToken: token,
	}

	return tokenAuth
}

// Authenticate authenticates with Vault and returns a token
func (t *TokenAuth) Authenticate(vaultClient *api.Client) error {

	// If we cannot write the Vault token, we'll just have to login next time. Nothing showstopping.
	err := utils.SetToken(vaultClient, t.AccessToken)
	if err != nil {
		print(err)
	}

	return nil
}
