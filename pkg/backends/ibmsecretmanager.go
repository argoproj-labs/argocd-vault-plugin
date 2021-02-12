package backends

import (
	"fmt"

	"github.com/IBM/argocd-vault-plugin/pkg/types"
	"github.com/hashicorp/vault/api"
)

// IBMSecretManager is a struct for working with IBM Secret Manager
type IBMSecretManager struct {
	types.AuthType
	IBMCloudAPIKey string
	VaultClient    *api.Client
}

// New TODO
func NewIBMSecretManagerBackend(authType types.AuthType, client *api.Client) *IBMSecretManager {
	ibmSecretManager := &IBMSecretManager{
		AuthType:    authType,
		VaultClient: client,
	}

	return ibmSecretManager
}

// Login authenticates with IBM Cloud Secret Manager using IAM and returns a token
func (i *IBMSecretManager) Login() error {
	err := i.AuthType.Authenticate(i.VaultClient)
	if err != nil {
		return err
	}
	return nil
}

// GetSecrets gets secrets from IBM Secret Manager and returns the formatted data
func (i *IBMSecretManager) GetSecrets(path, _ string) (map[string]interface{}, error) {
	secret, err := i.VaultClient.Logical().Read(path)
	if err != nil {
		return nil, err
	}

	if secret == nil {
		return nil, fmt.Errorf("Could not find secrets at path %s", path)
	}

	var data map[string]interface{}
	data = secret.Data

	// Make sure the secret exists
	if _, ok := data["secrets"]; !ok {
		return nil, fmt.Errorf("Could not find secrets at path %s", path)
	}

	// Get list of secrets
	secretList := data["secrets"].([]interface{})
	v := make([]string, 0, len(secretList))
	// Loop through secrets and get id
	// as getting the list of secrets does not include the payload
	for _, value := range secretList {
		secret := value.(map[string]interface{})
		if t, found := secret["id"]; found {
			v = append(v, t.(string))
		}
	}

	// Read each secret and get payload
	secrets := make(map[string]interface{})
	for _, j := range v {
		secret, err := i.VaultClient.Logical().Read(fmt.Sprintf("%s/%s", path, j))
		if err != nil {
			return nil, err
		}

		var data map[string]interface{}
		data = secret.Data

		// Get name and data of secret and append to secrets map
		secretName := data["name"].(string)
		secretData := data["secret_data"].(map[string]interface{})
		secrets[secretName] = secretData["payload"]
	}

	return secrets, nil
}
