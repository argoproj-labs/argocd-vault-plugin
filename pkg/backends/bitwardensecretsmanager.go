package backends

import (
	"fmt"

	bitwarden "github.com/bitwarden/sdk-go"
)

type BitwardenSecretsClient interface {
	List(string) (*bitwarden.SecretIdentifiersResponse, error)
	Get(string) (*bitwarden.SecretResponse, error)
}

// BitwardenSecretsManager is a struct for working with a Bitwarden Secrets Manager backend
type BitwardenSecretsManager struct {
	Client BitwardenSecretsClient
}

// NewBitwardenSecretsClient initializes a new Bitwarden Secrets Manager backend
func NewBitwardenSecretsClient(client BitwardenSecretsClient) *BitwardenSecretsManager {
	return &BitwardenSecretsManager{
		Client: client,
	}
}

// Login does nothing as a "login" is handled on the instantiation of the Bitwarden SDK
func (bw *BitwardenSecretsManager) Login() error {
	return nil
}

// GetSecrets gets secrets from Bitwarden Secrets Manager and returns the formatted data
// The path is of format `organization-id
func (bw *BitwardenSecretsManager) GetSecrets(path string, _ string, _ map[string]string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	secrets, err := bw.Client.List(path)
	if err != nil {
		return nil, err
	}

	for _, secret := range secrets.Data {
		value, err := bw.Client.Get(secret.ID)
		if err != nil {
			return nil, err
		}
		result[secret.ID] = value.Value
	}

	return result, nil
}

// GetIndividualSecret will get the specific secret (placeholder) from the SM backend
// The path is of format `organization-id/secret-id`
// organization id is ignored for indvidual secret fetching, but is included here to
// keep a standard path.
// Version is not supported and is ignored.
func (bw *BitwardenSecretsManager) GetIndividualSecret(_, secret, _ string, _ map[string]string) (interface{}, error) {
	fmt.Println(secret)
	value, err := bw.Client.Get(secret)
	if err != nil {
		return nil, err
	}
	return value.Value, nil
}
