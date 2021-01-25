package vault

// AppRole is a struct for working with Vault that uses AppRole
type AppRole struct {
	RoleID   string
	SecretID string
	*Client
}

// Login authenticates with Vault using App Role and returns a token
func (a *AppRole) Login() error {
	payload := map[string]interface{}{
		"role_id":   a.RoleID,
		"secret_id": a.SecretID,
	}

	data, err := a.Client.Write("auth/approle/login", payload)
	if err != nil {
		return err
	}

	// If we cannot write the Vault token, we'll just have to login next time. Nothing showstopping.
	err = SetToken(a.Client, data.Auth.ClientToken)
	if err != nil {
		print(err)
	}

	return nil
}

// GetSecrets gets secrets from vault and returns the formatted data
func (a *AppRole) GetSecrets(path string) (map[string]interface{}, error) {
	data, err := a.Client.Read(path)
	if err != nil {
		return nil, err
	}

	return data, nil
}
