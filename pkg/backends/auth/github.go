package auth

import (
	"github.com/IBM/argocd-vault-plugin/pkg/utils"
	"github.com/hashicorp/vault/api"
)

// Github is a struct for working with Vault that uses the Github Auth method
type Github struct {
	AccessToken string
}

// Authenticate authenticates with Vault and returns a token
func (g *Github) Authenticate(vaultClient *api.Client) error {
	payload := map[string]interface{}{
		"token": g.AccessToken,
	}

	data, err := vaultClient.Logical().Write("auth/github/login", payload)
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
