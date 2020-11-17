package vault

import "github.com/IBM/argocd-vault-plugin/pkg/vault/client"

// SecretManager is a struct for working with IBM Secret Manager
type SecretManager struct {
	IAMToken string
	token    string
	*client.VaultClient
}

// Login authenticates with IBM Cloud Secret Manager using IAM and returns a token
func (s *SecretManager) Login() error {
	payload := map[string]interface{}{
		"token": s.IAMToken,
	}

	data, err := s.VaultClient.Write("auth/ibmcloud/login", payload)
	if err != nil {
		return err
	}

	s.token = data.Auth.ClientToken
	return nil
}

// GetSecrets gets secrets from IBM Secret Manager and returns the formatted data
func (s *SecretManager) GetSecrets(path string) (map[string]interface{}, error) {
	data, err := s.VaultClient.Read(s.VaultClient.PathPrefix+path, s.token)
	if err != nil {
		return nil, err
	}

	// Will need to do more to handle the data

	return data, nil
}
