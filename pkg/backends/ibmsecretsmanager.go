package backends

import (
	"regexp"
	"fmt"

	ibmsm "github.com/IBM/secrets-manager-go-sdk/secretsmanagerv1"
	"github.com/IBM/go-sdk-core/v5/core"
)

var IBMPlaceholder, _ = regexp.Compile(`ibmcloud/(?P<type>.+)/secrets/groups/(?P<groupid>.+)`)

// IBMSecretsManagerClient is an interface for any client to the IBM Secrets Manager
// These are only the methods we need
type IBMSecretsManagerClient interface {
	ListAllSecrets(listAllSecretsOptions *ibmsm.ListAllSecretsOptions) (result *ibmsm.ListSecrets, response *core.DetailedResponse, err error)
	GetSecret(getSecretOptions *ibmsm.GetSecretOptions) (result *ibmsm.GetSecret, response *core.DetailedResponse, err error)
}

// IBMSecretsManager is a struct for working with IBM Secret Manager
type IBMSecretsManager struct {
	Client IBMSecretsManagerClient
}

// NewIBMSecretsManagerBackend initializes a new IBM Secret Manager backend
func NewIBMSecretsManagerBackend(client IBMSecretsManagerClient) *IBMSecretsManager {
	ibmSecretsManager := &IBMSecretsManager{
		Client: client,
	}
	return ibmSecretsManager
}

// Login does nothing since the IBM Secrets Manager client is setup on instantiation
func (i *IBMSecretsManager) Login() error {
	return nil
}

// GetSecrets gets secrets from IBM Secret Manager and returns the formatted data
func (i *IBMSecretsManager) GetSecrets(path string, _ map[string]string) (map[string]interface{}, error) {
	// 	secret, err := i.VaultClient.Logical().Read(path)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	if secret == nil {
	// 		return nil, fmt.Errorf("Could not find secrets at path %s", path)
	// 	}

	// 	var data map[string]interface{}
	// 	data = secret.Data

	// 	// Make sure the secret exists
	// 	if _, ok := data["secrets"]; !ok {
	// 		return nil, fmt.Errorf("Could not find secrets at path %s", path)
	// 	}

	// 	// Get list of secrets
	// 	secretList := data["secrets"].([]interface{})
	// 	v := make([]string, 0, len(secretList))
	// 	// Loop through secrets and get id
	// 	// as getting the list of secrets does not include the payload
	// 	for _, value := range secretList {
	// 		secret := value.(map[string]interface{})
	// 		if t, found := secret["id"]; found {
	// 			v = append(v, t.(string))
	// 		}
	// 	}

	// 	// Read each secret and get payload
	// 	secrets := make(map[string]interface{})
	// 	for _, j := range v {
	// 		secret, err := i.VaultClient.Logical().Read(fmt.Sprintf("%s/%s", path, j))
	// 		if err != nil {
	// 			return nil, err
	// 		}

	// 		if secret == nil || len(secret.Data) == 0 {
	// 			continue
	// 		}

	// 		var data map[string]interface{}
	// 		data = secret.Data

	// 		// Get name and data of secret and append to secrets map
	// 		secretName := data["name"].(string)
	// 		secretData := data["secret_data"].(map[string]interface{})
	// 		secrets[secretName] = secretData["payload"]
	// 	}

	// 	return secrets, nil
	return nil, nil
}

// GetSecretsVersioned returns the data for a secret in IBM Secrets Manager
// It only works for `arbitrary` secret types
func (i *IBMSecretsManager) GetSecretsVersioned(path string, version string, annotations map[string]string) (map[string]interface{}, error) {
	// IBM SM users pass the path of a secret _group_ which contains a list of secrets
	// ex: <path:ibmcloud/arbitrary/secrets/groups/123#username>
	// So we query the group to enumerate the secret ids, and retrieve each one to return a complete map of them
	matches := IBMPlaceholder.FindStringSubmatch(path)
	if len(matches) == 0 {
		return nil, fmt.Errorf("Path is not in the correct format (ibmcloud/$TYPE/secrets/groups/$GROUP_ID) for IBM Secrets Manager: %s", path)
	}

	// Enumerate the secret names and their ids
	groupid := matches[IBMPlaceholder.SubexpIndex("groupid")]
	result, details, err := i.Client.ListAllSecrets(&ibmsm.ListAllSecretsOptions{
		Groups: []string{groupid},
	})
	
	if err != nil {
		return nil, fmt.Errorf("Could not list secrets for secret group %s: %s\n%s", groupid, err, details)
	}

	secrets := make(map[string]interface{})
	for _, resource := range result.Resources {
		if secret, ok := resource.(*ibmsm.SecretResource); ok {
			if *secret.SecretType == matches[IBMPath.SubexpIndex("type")] {
				secrets[*secret.Name] = secret.ID
			}
		}
	}

	// Get each secrets value from its ID
	for name, id := range(secrets) {

		// `version` is ignored since IBM SM does not support versioning for `arbitrary` secrets
		// https://github.com/IBM/argocd-vault-plugin/issues/58#issuecomment-906477921
		secretRes, _, err := i.Client.GetSecret(&ibmsm.GetSecretOptions{
			SecretType: &matches[IBMPlaceholder.SubexpIndex("type")],
			ID:         id.(*string),
		})
		if err != nil {
			return nil, fmt.Errorf("Could not retrieve secret %s: %s", *(id.(*string)), err)
		}

		secretResource := secretRes.Resources[0].(*ibmsm.SecretResource)
		secretData := secretResource.SecretData.(map[string]interface{})
		secrets[name] = secretData["payload"]
		if secrets[name] == nil {
			return nil, fmt.Errorf("No `payload` key present for secret at path %s: Is this an `arbitrary` type secret?", path)
		}
	}

	return secrets, nil
}
