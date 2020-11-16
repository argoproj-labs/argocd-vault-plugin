package vault

import (
	vault "github.com/IBM/argocd-vault-plugin/pkg/vault/client"
)

// SecretManager is a struct for working with IBM Secret Manager
type SecretManager struct {
	IAMToken string
	token    string
}

// Login authenticates with IBM Cloud Secret Manager using IAM and returns a token
func (s *SecretManager) Login() error {
	client, _ := vault.NewVaultClient()

	payload := map[string]interface{}{
		"token": s.IAMToken,
	}

	data, err := client.Write("auth/ibmcloud/login", payload)
	if err != nil {
		return err
	}

	s.token = data.Auth.ClientToken
	return nil
}

// GetSecrets gets secrets from IBM Secret Manager and returns the formatted data
func (s *SecretManager) GetSecrets(path string) (map[string]interface{}, error) {
	client, _ := vault.NewVaultClient()

	data, err := client.Read(path, s.token)
	if err != nil {
		return nil, err
	}

	// Will need to do more to handle the data

	return data, nil
}
