package vault

import (
	"github.com/hashicorp/vault/api"
)

// Client is used to make API calls to Vault in a standard way
type Client struct {
	VaultAPIClient *api.Client
}

func (c *Client) Write(path string, payload map[string]interface{}) (*api.Secret, error) {
	data, err := c.VaultAPIClient.Logical().Write(path, payload)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *Client) Read(path string) (map[string]interface{}, error) {
	data, err := c.VaultAPIClient.Logical().Read(path)
	if err != nil {
		return nil, err
	}

	if data != nil {
		return data.Data, nil
	}
	// if nil throw error or empty map
	return map[string]interface{}{}, nil
}
