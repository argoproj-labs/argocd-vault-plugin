package vault

// SecretManager is a struct for working with IBM Secret Manager
type SecretManager struct {
	IAMToken string
	*Client
}

// Login authenticates with IBM Cloud Secret Manager using IAM and returns a token
func (s *SecretManager) Login() error {
	payload := map[string]interface{}{
		"token": s.IAMToken,
	}

	data, err := s.Client.Write("auth/ibmcloud/login", payload)
	if err != nil {
		return err
	}

	s.Client.VaultAPIClient.SetToken(data.Auth.ClientToken)
	return nil
}

// GetSecrets gets secrets from IBM Secret Manager and returns the formatted data
func (s *SecretManager) GetSecrets(path string) (map[string]interface{}, error) {
	data, err := s.Client.Read(s.Client.PathPrefix + path)
	if err != nil {
		return nil, err
	}

	// Will need to do more to handle the data

	return data, nil
}
