package vault

import "github.com/IBM/argocd-vault-plugin/pkg/vault/client"

// Github is a struct for working with Vault that uses the Github Auth method
type Github struct {
	AccessToken string
	token       string
	*client.VaultClient
}

// Login authenticates with Vault and returns a token
func (g *Github) Login() error {
	payload := map[string]interface{}{
		"token": g.AccessToken,
	}

	data, err := g.VaultClient.Write("auth/github/login", payload)
	if err != nil {
		return err
	}

	g.token = data.Auth.ClientToken
	return nil
}

// GetSecrets gets secrets from vault and returns the formatted data
func (g *Github) GetSecrets(path string) (map[string]interface{}, error) {
	data, err := g.VaultClient.Read(path, g.token)
	if err != nil {
		return nil, err
	}

	return data, nil
}
