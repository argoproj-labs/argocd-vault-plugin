package backends

import (
	"fmt"
	"regexp"

	"github.com/IBM/go-sdk-core/v5/core"
	ibmsm "github.com/IBM/secrets-manager-go-sdk/secretsmanagerv1"
)

var IBMPath, _ = regexp.Compile(`ibmcloud/(?P<type>.+)/secrets/groups/(?P<groupid>.+)`)

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

// GetSecrets returns the data for a secret in IBM Secrets Manager
// It only works for `arbitrary` secret types
func (i *IBMSecretsManager) GetSecrets(path string, version string, annotations map[string]string) (map[string]interface{}, error) {
	// IBM SM users pass the path of a secret _group_ which contains a list of secrets
	// ex: <path:ibmcloud/arbitrary/secrets/groups/123#username>
	// So we query the group to enumerate the secret ids, and retrieve each one to return a complete map of them
	matches := IBMPath.FindStringSubmatch(path)
	if len(matches) == 0 {
		return nil, fmt.Errorf("Path is not in the correct format (ibmcloud/$TYPE/secrets/groups/$GROUP_ID) for IBM Secrets Manager: %s", path)
	}

	// Enumerate the secret names and their ids
	groupid := matches[IBMPath.SubexpIndex("groupid")]
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
	for name, id := range secrets {

		// `version` is ignored since IBM SM does not support versioning for `arbitrary` secrets
		// https://github.com/IBM/argocd-vault-plugin/issues/58#issuecomment-906477921
		secretRes, _, err := i.Client.GetSecret(&ibmsm.GetSecretOptions{
			SecretType: &matches[IBMPath.SubexpIndex("type")],
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
