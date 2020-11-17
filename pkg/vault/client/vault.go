package client

import (
	"github.com/hashicorp/vault/api"
)

// VaultClient is used to make API calls to Vault in a standard way
type VaultClient struct {
	VaultAPIClient *api.Client
}

func (v *VaultClient) Write(path string, payload map[string]interface{}) (*api.Secret, error) {
	data, err := v.VaultAPIClient.Logical().Write(path, payload)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (v *VaultClient) Read(path string, token string) (map[string]interface{}, error) {
	v.VaultAPIClient.SetToken(token)
	data, err := v.VaultAPIClient.Logical().Read(path)
	if err != nil {
		return nil, err
	}

	return data.Data, nil
}
