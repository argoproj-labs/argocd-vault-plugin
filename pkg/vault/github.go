package vault

// Github is a struct for working with Vault that uses the Github Auth method
type Github struct {
	AccessToken string
	*Client
}

// Login authenticates with Vault and returns a token
func (g *Github) Login() error {
	payload := map[string]interface{}{
		"token": g.AccessToken,
	}

	data, err := g.Client.Write("auth/github/login", payload)
	if err != nil {
		return err
	}

	g.Client.VaultAPIClient.SetToken(data.Auth.ClientToken)
	return nil
}

// GetSecrets gets secrets from vault and returns the formatted data
func (g *Github) GetSecrets(path string) (map[string]interface{}, error) {
	data, err := g.Client.Read(path)
	if err != nil {
		return nil, err
	}

	return data, nil
}
