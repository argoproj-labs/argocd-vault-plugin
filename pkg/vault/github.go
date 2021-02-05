package vault

// Github is a struct for working with Vault that uses the Github Auth method
type Github struct {
	AccessToken string
	KvVersion   string
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

	// If we cannot write the Vault token, we'll just have to login next time. Nothing showstopping.
	err = SetToken(g.Client, data.Auth.ClientToken)
	if err != nil {
		print(err)
	}

	return nil
}

// GetSecrets gets secrets from vault and returns the formatted data
func (g *Github) GetSecrets(path, kvVersion string) (map[string]interface{}, error) {
	if kvVersion != "" {
		g.KvVersion = kvVersion
	}
	return ReadVaultSecret(*g.Client, path, g.KvVersion)
}
