package backends

import (
	"strings"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
)

// OnePassword is a struct for working with a 1Password Connect backend
type OnePasswordConnect struct {
	Client connect.Client
}

// NewOnePasswordConnectBackend initializes a new 1Password Connect backend
func NewOnePasswordConnectBackend(client connect.Client) *OnePasswordConnect {
	return &OnePasswordConnect{
		Client: client,
	}
}

// Login does nothing as a "login" is handled on the instantiation of the 1Password Connect SDK
func (a *OnePasswordConnect) Login() error {
	return nil
}

// GetSecrets gets secrets from 1Password Connect server and returns the formatted data
func (a *OnePasswordConnect) GetSecrets(path string, version string, annotations map[string]string) (map[string]interface{}, error) {
	// Format we expect is vaults/<vaultUUID>/items/<secret_UUID>
	splits := strings.Split(path, "/")
	vaultUUID := splits[1]
	itemUUID := splits[3]

	utils.VerboseToStdErr("OnePassword Connect getting item %s from vault %s", itemUUID, vaultUUID)
	result, err := a.Client.GetItem(itemUUID, vaultUUID)
	if err != nil {
		return nil, err
	}

	utils.VerboseToStdErr("OnePassword Connect get secret response: %v", result)

	data := make(map[string]interface{})

	for _, field := range result.Fields {
		data[field.Label] = field.Value
	}

	return data, nil
}

// GetIndividualSecret will get the specific secret (placeholder) from the 1Password connect backend
// For 1Password, we only support placeholders replaced from the k/v pairs of a secret which cannot be individually addressed
// So, we use GetSecrets and extract the specific placeholder we want
func (a *OnePasswordConnect) GetIndividualSecret(kvpath, secret, version string, annotations map[string]string) (interface{}, error) {
	data, err := a.GetSecrets(kvpath, version, annotations)
	if err != nil {
		return nil, err
	}
	return data[secret], nil
}
