package vault

import (
	"github.com/IBM/argocd-vault-plugin/pkg/utils"
	"github.com/hashicorp/vault/api"
)

// GithubAuth is a struct for working with Vault that uses the Github Auth method
type GithubAuth struct {
	AccessToken string
}

// NewGithubAuth initializes a new GithubAuth with token
func NewGithubAuth(token string) *GithubAuth {
	githubAuth := &GithubAuth{
		AccessToken: token,
	}

	return githubAuth
}

// Authenticate authenticates with Vault and returns a token
func (g *GithubAuth) Authenticate(vaultClient *api.Client) error {
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
