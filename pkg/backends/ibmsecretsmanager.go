package backends

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/IBM/argocd-vault-plugin/pkg/types"
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

// getSecret sends the result of getting the `secret` from IBM SM in a map over a channel
// `name` is the name of the secret and is always set
// `err` is set if there is an error getting the secret
// `payload` is the secrets `payload` and is set if successful
// The goroutine only terminates once IBMMaxRetries or fewer attempts are made
func (i *IBMSecretsManager) getSecret(secret *ibmsm.SecretResource, response chan map[string]interface{}, wg *sync.WaitGroup) {
	result := make(map[string]interface{})
	result["name"] = *secret.Name

	// `version` is ignored since IBM SM does not support versioning for `arbitrary` secrets
	// https://github.com/IBM/argocd-vault-plugin/issues/58#issuecomment-906477921
	secretRes, httpResponse, err := i.Client.GetSecret(&ibmsm.GetSecretOptions{
		SecretType: secret.SecretType,
		ID:         secret.ID,
	})
	if err != nil {
		result["err"] = fmt.Errorf("Could not retrieve secret %s: %s", *secret.ID, err)
	}

	if secretRes == nil {
		result["err"] = fmt.Errorf("Could not retrieve secret %s after %d retries, statuscode %d", *secret.ID, types.IBMMaxRetries, httpResponse.GetStatusCode())
	} else {
		secretResource := secretRes.Resources[0].(*ibmsm.SecretResource)

		// IAM credentials do not come back in the `secretData`
		if *secretResource.SecretType == types.IBMIAMCredentialsType {
			result["payload"] = map[string]interface{}{
				"api_key": *secretResource.APIKey,
			}
		} else {
			secretData := secretResource.SecretData.(map[string]interface{})

			// Copy whatever keys this non-arbitrary secret has into a map for use with `jsonParse`
			if secretData["payload"] == nil {
				result["payload"] = make(map[string]interface{})
				for k, v := range secretData {
					(result["payload"].(map[string]interface{}))[k] = v
				}
			} else {
				result["payload"] = secretData["payload"]
			}
		}
	}

	response <- result
	wg.Done()
}

func storeSecret(secrets *map[string]interface{}, result map[string]interface{}) error {
	if result["err"] != nil {
		return result["err"].(error)
	}
	(*secrets)[result["name"].(string)] = result["payload"]
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
	// The IBM SM API returns a max of MAX_PER_PAGE results, so if we get that many on the first request, there might be more secrets
	groupid := matches[IBMPath.SubexpIndex("groupid")]
	var offset int64 = 0
	var result []ibmsm.SecretResourceIntf
	for {
		res, details, err := i.Client.ListAllSecrets(&ibmsm.ListAllSecretsOptions{
			Groups: []string{groupid},
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("Could not list secrets for secret group %s: %s\n%s", groupid, err, details.String())
		}
		if res == nil {
			return nil, fmt.Errorf("Could not list secrets for secret group %s: %d\n%s", groupid, details.GetStatusCode(), details.String())
		}

		result = append(result, res.Resources...)

		if len(res.Resources) < types.IBMMaxPerPage {
			break
		}
		offset += int64(types.IBMMaxPerPage)
	}

	// Using MAX_GOROUTINES at a time, retrieve the secrets of the right type from the group
	secretResult := make(chan map[string]interface{})
	secrets := make(map[string]interface{})
	var wg sync.WaitGroup
	MAX_GOROUTINES := 20
	launchedRoutines := 0

	for _, resource := range result {
		if secret, ok := resource.(*ibmsm.SecretResource); ok {

			// This check is required since secrets are only unique by name, group, _and_ type
			if *secret.SecretType == matches[IBMPath.SubexpIndex("type")] {

				// There is space for more goroutines, so spawn immediately and continue
				if launchedRoutines < MAX_GOROUTINES {
					go i.getSecret(secret, secretResult, &wg)
					wg.Add(1)
					launchedRoutines += 1
					continue
				}

				// Wait for a goroutine to finish before spawning another
				err := storeSecret(&secrets, <-secretResult)
				if err != nil {
					return nil, err
				}

				go i.getSecret(secret, secretResult, &wg)
				wg.Add(1)
				launchedRoutines += 1
			}
		}
	}

	go func() {
		wg.Wait()
		close(secretResult)
	}()

	for secret := range secretResult {
		err := storeSecret(&secrets, secret)
		if err != nil {
			return nil, err
		}
	}

	return secrets, nil
}

// GetIndividualSecret will get the specific secret (placeholder) from the SM backend
// For IBM, we only support placeholders replaced from arbitrary secrets in a group, which cannot be individually addressed by placeholder
// So, we use GetSecrets and extract the specific placeholder we want
func (i *IBMSecretsManager) GetIndividualSecret(kvpath, secret, version string, annotations map[string]string) (interface{}, error) {
	data, err := i.GetSecrets(kvpath, version, annotations)
	if err != nil {
		return nil, err
	}
	return data[secret], nil
}
