package client

import (
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/vault/api"
)

// Client is used to make API calls to Vault in a standard way
type Client struct {
	client *api.Client
}

// NewVaultClient initilizes an instance of VaultClient
func NewVaultClient() (*Client, error) {
	var httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	client, err := api.NewClient(&api.Config{Address: os.Getenv("VAULT_ADDR"), HttpClient: httpClient})
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}

func (v *Client) Write(path string, payload map[string]interface{}) (*api.Secret, error) {
	data, err := v.client.Logical().Write(path, payload)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (v *Client) Read(path string, token string) (map[string]interface{}, error) {
	v.client.SetToken(token)
	data, err := v.client.Logical().Read(path)
	if err != nil {
		return nil, err
	}

	return data.Data, nil
}
