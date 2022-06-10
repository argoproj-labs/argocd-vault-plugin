package backends

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/IBM/go-sdk-core/v5/core"
	ibmsm "github.com/IBM/secrets-manager-go-sdk/secretsmanagerv1"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/types"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
)

var IBMPath, _ = regexp.Compile(`ibmcloud/(?P<type>.+)/secrets/groups/(?P<groupId>.+)`)

// IBMSecretsManagerClient is an interface for any client to the IBM Secrets Manager
// These are only the methods we need
type IBMSecretsManagerClient interface {
	ListAllSecrets(listAllSecretsOptions *ibmsm.ListAllSecretsOptions) (result *ibmsm.ListSecrets, response *core.DetailedResponse, err error)
	GetSecret(getSecretOptions *ibmsm.GetSecretOptions) (result *ibmsm.GetSecret, response *core.DetailedResponse, err error)
	GetSecretVersion(getSecretOptions *ibmsm.GetSecretVersionOptions) (result *ibmsm.GetSecretVersion, response *core.DetailedResponse, err error)
}

// Used as the key into the several caches for IBM SM API calls
// Includes groupId and secretType since secrets are unique by group, type, and their name
type cacheKey struct {
	groupId    string
	secretType string
}

// IBMSecretsManager is a struct for working with IBM Secret Manager
type IBMSecretsManager struct {
	Client IBMSecretsManagerClient

	// Cache for storing *ibmsm.SecretResource's from listing the secrets of a group
	// Organized as:
	// [groupId]: { [secretType]: { [secretName]: &ibmsm.SecretResource } }
	// Only read/written to by the main goroutine, no synchronized access needed
	listAllSecretsCache map[cacheKey]map[string]*ibmsm.SecretResource

	// Cache for storing payloads (interface{}) of secrets
	// Organized as:
	// [groupId]: { [secretType]: { [secretName]: interface{} } }
	// We don't keep track of the secret version since most secrets aren't versionable in IBM SM anyway,
	// so this cache should not be used to retrieve a secret with a specific version
	// Written to by the `i.getSecrets` goroutines, synchronized access provided by getSecretsCacheLock
	getSecretsCache map[cacheKey]map[string]interface{}

	getSecretsCacheLock sync.RWMutex

	// Keeps track of whether GetSecrets has been called for a given group and secret type
	// Only read/written to by the main goroutine, no synchronized access needed
	retrievedAllSecrets map[cacheKey]bool
}

// NewIBMSecretsManagerBackend initializes a new IBM Secret Manager backend
func NewIBMSecretsManagerBackend(client IBMSecretsManagerClient) *IBMSecretsManager {
	ibmSecretsManager := &IBMSecretsManager{
		Client:              client,
		listAllSecretsCache: make(map[cacheKey]map[string]*ibmsm.SecretResource),
		getSecretsCache:     make(map[cacheKey]map[string]interface{}),
		retrievedAllSecrets: make(map[cacheKey]bool),
	}
	return ibmSecretsManager
}

// parsePath returns the groupId, secretType represented by the path
func parsePath(path string) (string, string, error) {
	matches := IBMPath.FindStringSubmatch(path)
	if len(matches) == 0 {
		return "", "", fmt.Errorf("Path is not in the correct format (ibmcloud/$TYPE/secrets/groups/$GROUP_ID) for IBM Secrets Manager: %s", path)
	}
	return matches[IBMPath.SubexpIndex("type")], matches[IBMPath.SubexpIndex("groupId")], nil
}

func (i *IBMSecretsManager) readSecretFromCache(groupId, secretType, secretName string) interface{} {
	result := i.getSecretsCache[cacheKey{groupId, secretType}]
	if result != nil {
		return result[secretName]
	}
	return nil
}

func (i *IBMSecretsManager) writeSecretToCache(groupId, secretType, secretName string, payload interface{}) {
	ckey := cacheKey{groupId, secretType}
	if i.getSecretsCache[ckey] != nil {
		i.getSecretsCache[ckey][secretName] = payload
	} else {
		i.getSecretsCache[ckey] = map[string]interface{}{
			secretName: payload,
		}
	}
}

// Login does nothing since the IBM Secrets Manager client is setup on instantiation
func (i *IBMSecretsManager) Login() error {
	return nil
}

// getSecretVersionedOrNot will ultimately return the payload of a secret from IBM SM:
// - `secret_data` map for arbitrary secrets
// - `api_key` k/v pair for IAM credential secrets
// - `certificate`, `private_key`, etc. k/v pairs for versioned certificate secrets
// API calls and their responses depend on the whether the secret "can be" versioned or not
func (i *IBMSecretsManager) getSecretVersionedOrNot(secret *ibmsm.SecretResource, version string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Only certificate secrets are versioned in IBM SM
	if version != "" && (*secret.SecretType == types.IBMImportedCertType || *secret.SecretType == types.IBMPublicCertType) {
		opts := &ibmsm.GetSecretVersionOptions{
			SecretType: secret.SecretType,
			ID:         secret.ID,
			VersionID:  &version,
		}

		secretVersion, httpResponse, err := i.Client.GetSecretVersion(opts)
		if err != nil {
			return nil, fmt.Errorf("Could not retrieve secret %s: %s", *secret.ID, err)
		}
		if secretVersion == nil {
			return nil, fmt.Errorf("Could not retrieve secret %s after %d retries, statuscode %d", *secret.ID, types.IBMMaxRetries, httpResponse.GetStatusCode())
		}

		utils.VerboseToStdErr("IBM Cloud Secrets Manager get versioned secret %s HTTP response: %v", *secret.ID, httpResponse)

		result = (secretVersion.Resources[0].(*ibmsm.SecretVersion)).SecretData.(map[string]interface{})
	} else {
		secretRes, httpResponse, err := i.Client.GetSecret(&ibmsm.GetSecretOptions{
			SecretType: secret.SecretType,
			ID:         secret.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("Could not retrieve secret %s: %s", *secret.ID, err)
		}
		if secretRes == nil {
			return nil, fmt.Errorf("Could not retrieve secret %s after %d retries, statuscode %d", *secret.ID, types.IBMMaxRetries, httpResponse.GetStatusCode())
		}

		utils.VerboseToStdErr("IBM Cloud Secrets Manager get unversioned secret %s HTTP response: %v", *secret.ID, httpResponse)

		// APIKey secrets don't come from `SecretData`
		if *secret.SecretType == types.IBMIAMCredentialsType {
			result = map[string]interface{}{
				"api_key": *secretRes.Resources[0].(*ibmsm.SecretResource).APIKey,
			}
		} else {
			result = secretRes.Resources[0].(*ibmsm.SecretResource).SecretData.(map[string]interface{})
		}
	}

	return result, nil
}

// getSecret sends the result of getting the `secret` from IBM SM in a map over a channel
// `name` is the name of the secret and is always set
// `err` is set if there is an error getting the secret
// `payload` is the secrets `payload` and is set if successful
// The goroutine only terminates once IBMMaxRetries or fewer attempts are made
func (i *IBMSecretsManager) getSecret(secret *ibmsm.SecretResource, version string, response chan map[string]interface{}, wg *sync.WaitGroup) {
	result := make(map[string]interface{})
	result["name"] = *secret.Name

	var groupId string
	if secret.SecretGroupID == nil {
		groupId = "default"
	} else {
		groupId = *(secret.SecretGroupID)
	}

	secretType := *(secret.SecretType)
	secretName := *(secret.Name)

	i.getSecretsCacheLock.RLock()
	cacheResult := i.readSecretFromCache(groupId, secretType, secretName)
	i.getSecretsCacheLock.RUnlock()

	// Bypass the cache when explicit version is requested
	if cacheResult != nil && version == "" {
		utils.VerboseToStdErr("IBM Cloud Secrets Manager get secret: cache hit for %s of type %s from group %s", secretName, secretType, groupId)
		result["payload"] = cacheResult
	} else {
		utils.VerboseToStdErr("IBM Cloud Secrets Manager get secret: getting secret %s of type %s from group %s", secretName, secretType, groupId)
		secretData, err := i.getSecretVersionedOrNot(secret, version)
		var payload interface{}
		if err != nil {
			result["err"] = err
		} else {

			// Copy whatever keys this non-arbitrary secret has into a map for use with `jsonParse`
			if secretData["payload"] == nil {
				payload = make(map[string]interface{})
				for k, v := range secretData {
					(payload.(map[string]interface{}))[k] = v
				}
			} else {
				payload = secretData["payload"]
			}
		}

		// Populate cache if successful
		if err == nil {
			i.getSecretsCacheLock.Lock()
			i.writeSecretToCache(groupId, secretType, secretName, payload)
			i.getSecretsCacheLock.Unlock()
		}

		result["payload"] = payload
	}

	response <- result
	wg.Done()
}

// Enumerate the secret names and their ids for the secrets of type secretType in group groupId,
// caching results into listAllSecretsCache
func (i *IBMSecretsManager) listSecretsInGroup(groupId, secretType string) (map[string]*ibmsm.SecretResource, error) {
	ckey := cacheKey{groupId, secretType}
	cachedData := i.listAllSecretsCache[ckey]
	if cachedData != nil {
		utils.VerboseToStdErr("IBM Cloud Secrets Manager list secrets in group: cache hit group %s", groupId)
		return cachedData, nil
	}

	var offset int64 = 0
	for {
		utils.VerboseToStdErr("IBM Cloud Secrets Manager listing secrets of from group %s starting at offset %d", groupId, offset)
		res, details, err := i.Client.ListAllSecrets(&ibmsm.ListAllSecretsOptions{
			Groups: []string{groupId},
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("Could not list secrets for secret group %s: %s\n%s", groupId, err, details.String())
		}
		if res == nil {
			return nil, fmt.Errorf("Could not list secrets for secret group %s: %d\n%s", groupId, details.GetStatusCode(), details.String())
		}

		utils.VerboseToStdErr("IBM Cloud Secrets Manager list secrets in group HTTP response: %v", details)

		for _, secret := range res.Resources {
			name := *(secret.(*ibmsm.SecretResource).Name)
			ttype := *(secret.(*ibmsm.SecretResource).SecretType)
			ckey := cacheKey{groupId, ttype}

			if i.listAllSecretsCache[ckey] != nil {
				i.listAllSecretsCache[ckey][name] = secret.(*ibmsm.SecretResource)
			} else {
				i.listAllSecretsCache[ckey] = map[string]*ibmsm.SecretResource{
					name: secret.(*ibmsm.SecretResource),
				}
			}
		}

		// The IBM SM API returns a max of MAX_PER_PAGE results, so if we get that many on the first request, there might be more secrets
		if len(res.Resources) < types.IBMMaxPerPage {
			break
		}
		offset += int64(types.IBMMaxPerPage)
	}

	return i.listAllSecretsCache[ckey], nil
}

func storeSecret(secrets *map[string]interface{}, result map[string]interface{}) error {
	if result["err"] != nil {
		return result["err"].(error)
	}
	(*secrets)[result["name"].(string)] = result["payload"]
	return nil
}

// GetSecrets returns the data for all secrets of a specific type of a group in IBM Secrets Manager
func (i *IBMSecretsManager) GetSecrets(path string, version string, annotations map[string]string) (map[string]interface{}, error) {
	secretType, groupId, err := parsePath(path)
	if err != nil {
		return nil, fmt.Errorf("Path is not in the correct format (ibmcloud/$TYPE/secrets/groups/$GROUP_ID) for IBM Secrets Manager: %s", path)
	}
	ckey := cacheKey{groupId, secretType}

	// Bypass the cache when explicit version is requested
	// Otherwise, use it if applicable
	if version == "" && i.retrievedAllSecrets[ckey] {
		return i.getSecretsCache[ckey], nil
	}

	// So we query the group to enumerate the secret ids, and retrieve each one to return a complete map of them
	result, err := i.listSecretsInGroup(groupId, secretType)
	if err != nil {
		return nil, err
	}

	// Using MAX_GOROUTINES at a time, retrieve the secrets of the right type from the group
	secretResult := make(chan map[string]interface{})
	secrets := make(map[string]interface{})
	var wg sync.WaitGroup
	MAX_GOROUTINES := 20
	launchedRoutines := 0

	for _, secret := range result {

		// There is space for more goroutines, so spawn immediately and continue
		if launchedRoutines < MAX_GOROUTINES {
			go i.getSecret(secret, version, secretResult, &wg)
			wg.Add(1)
			launchedRoutines += 1
			continue
		}

		// Wait for a goroutine to finish before spawning another
		err := storeSecret(&secrets, <-secretResult)
		if err != nil {
			return nil, err
		}

		go i.getSecret(secret, version, secretResult, &wg)
		wg.Add(1)
		launchedRoutines += 1
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

	i.retrievedAllSecrets[ckey] = true

	return secrets, nil
}

// GetIndividualSecret will get the specific secret (placeholder) from the SM backend
// This requires listing the secrets of the group to obtain the id, and then using that to grab the one secret's payload
func (i *IBMSecretsManager) GetIndividualSecret(kvpath, secretName, version string, annotations map[string]string) (interface{}, error) {
	secretType, groupId, err := parsePath(kvpath)
	if err != nil {
		return nil, fmt.Errorf("Path is not in the correct format (ibmcloud/$TYPE/secrets/groups/$GROUP_ID) for IBM Secrets Manager: %s", kvpath)
	}
	ckey := cacheKey{groupId, secretType}

	// Bypass the cache when explicit version is requested
	// If we have already retrieved all the secrets for the requested secret's group and type, we have a cache hit
	if version == "" && i.retrievedAllSecrets[ckey] {
		return i.getSecretsCache[ckey][secretName], nil
	}

	// Grab the *ibmsm.SecretResource corresponding to the secret
	secretResources, err := i.listSecretsInGroup(groupId, secretType)
	if err != nil {
		return nil, err
	}
	secret := secretResources[secretName]
	if secret == nil {

		// Allow the replacement code to handle this missing secret
		return nil, nil
	}

	// Retrieve the secret's payload
	secrets := make(map[string]interface{})
	secretResult := make(chan map[string]interface{})
	var wg sync.WaitGroup
	go i.getSecret(secret, version, secretResult, &wg)
	wg.Add(1)
	go func() {
		wg.Wait()
		close(secretResult)
	}()

	err = storeSecret(&secrets, <-secretResult)
	if err != nil {
		return nil, err
	}

	return secrets[secretName], nil
}
