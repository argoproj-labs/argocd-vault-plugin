package backends_test

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/IBM/go-sdk-core/v5/core"
	ibmsm "github.com/IBM/secrets-manager-go-sdk/secretsmanagerv2"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/backends"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/types"
)

type MockIBMSMClient struct {
	ListSecretsOptionCalledWith []*ibmsm.ListSecretsOptions

	// GetSecretLock prevents false data races caused by unsychronized access to the mock state
	// It is shared b/w both GetSecret and GetSecretVersion for simplicity, even though each writes to a different field
	GetSecretLock sync.RWMutex

	GetSecretCalledWith        *ibmsm.GetSecretOptions
	GetSecretCallCount         int
	GetSecretVersionCalledWith *ibmsm.GetSecretVersionOptions
	GetSecretVersionCallCount  int
}

var BIG_GROUP_LEN int = types.IBMMaxPerPage + 1

// This is used to take deep copies of struct fields passed as pointers in ListSecretsOptions
// so we can make assertions about the values later
func deepCopy(listAllSecretsOptions *ibmsm.ListSecretsOptions) *ibmsm.ListSecretsOptions {
	var offset int64 = *listAllSecretsOptions.Offset
	return &ibmsm.ListSecretsOptions{
		Groups: listAllSecretsOptions.Groups,
		Offset: &offset,
	}
}

func (m *MockIBMSMClient) ListSecrets(listAllSecretsOptions *ibmsm.ListSecretsOptions) (result *ibmsm.SecretMetadataPaginatedCollection, response *core.DetailedResponse, err error) {
	m.ListSecretsOptionCalledWith = append(m.ListSecretsOptionCalledWith, deepCopy(listAllSecretsOptions))

	// A big secret group
	bigGroup := "big-group"
	stype := "arbitrary"
	bigGroupSecrets := make([]ibmsm.SecretMetadataIntf, BIG_GROUP_LEN)
	for id := 0; id < BIG_GROUP_LEN; id += 1 {
		name := fmt.Sprintf("my-secret-%d", id)
		bigGroupSecrets[id] = &ibmsm.ArbitrarySecretMetadata{
			Name:          &name,
			SecretType:    &stype,
			SecretGroupID: &bigGroup,
			ID:            &stype,
		}
	}

	// A small secret  group
	smallGroup := "small-group"
	name := "my-secret"
	otype := "username_password"
	ctype := "public_cert"
	itype := "iam_credentials"
	ktype := "kv"
	smallGroupSecrets := []ibmsm.SecretMetadataIntf{
		&ibmsm.ArbitrarySecretMetadata{
			Name:          &name,
			SecretType:    &stype,
			SecretGroupID: &smallGroup,
			ID:            &stype,
		},
		&ibmsm.UsernamePasswordSecretMetadata{
			Name:          &name,
			SecretType:    &otype,
			SecretGroupID: &smallGroup,
			ID:            &otype,
		},
		&ibmsm.PublicCertificateMetadata{
			Name:          &name,
			SecretType:    &ctype,
			SecretGroupID: &smallGroup,
			ID:            &ctype,
		},
		&ibmsm.IAMCredentialsSecretMetadata{
			Name:          &name,
			SecretType:    &itype,
			SecretGroupID: &smallGroup,
			ID:            &itype,
		},
		&ibmsm.KVSecretMetadata{
			Name:          &name,
			SecretType:    &ktype,
			SecretGroupID: &smallGroup,
			ID:            &ktype,
		},
	}

	// Empty secret group
	emptyGroup := "empty-group"
	emptyGroupSecrets := []ibmsm.SecretMetadataIntf{}

	if listAllSecretsOptions.Groups[0] == bigGroup {
		// Emulate a 2-page paginated response
		offset := int(*listAllSecretsOptions.Offset)

		end := offset + types.IBMMaxPerPage
		if end > BIG_GROUP_LEN {
			end = BIG_GROUP_LEN
		}

		return &ibmsm.SecretMetadataPaginatedCollection{
			Secrets: bigGroupSecrets[offset:end],
		}, nil, nil
	} else if listAllSecretsOptions.Groups[0] == smallGroup {
		return &ibmsm.SecretMetadataPaginatedCollection{
			Secrets: smallGroupSecrets,
		}, nil, nil
	} else if listAllSecretsOptions.Groups[0] == emptyGroup {
		return &ibmsm.SecretMetadataPaginatedCollection{
			Secrets: emptyGroupSecrets,
		}, nil, nil
	} else {
		return nil, nil, fmt.Errorf("No such group %s", listAllSecretsOptions.Groups[0])
	}
}

func (m *MockIBMSMClient) GetSecret(getSecretOptions *ibmsm.GetSecretOptions) (result ibmsm.SecretIntf, response *core.DetailedResponse, err error) {
	m.GetSecretLock.Lock()
	m.GetSecretCalledWith = getSecretOptions
	m.GetSecretCallCount += 1
	m.GetSecretLock.Unlock()

	if *getSecretOptions.ID == "arbitrary" {
		name := "my-secret"
		id := "arbitrary"
		payload := "password"
		return &ibmsm.ArbitrarySecret{
			Name:    &name,
			ID:      &id,
			Payload: &payload,
		}, nil, nil
	} else if *getSecretOptions.ID == "iam_credentials" {
		name := "my-secret"
		id := "iam_credentials"
		payload := "password"
		return &ibmsm.IAMCredentialsSecret{
			Name:   &name,
			ID:     &id,
			ApiKey: &payload,
		}, nil, nil
	} else if *getSecretOptions.ID == "kv" {
		name := "my-secret"
		id := "kv"
		payload := map[string]interface{}{
			"hello": "there",
		}
		return &ibmsm.KVSecret{
			Name: &name,
			ID:   &id,
			Data: payload,
		}, nil, nil
	} else {
		name := "my-secret"
		id := "username_password"
		user := "user"
		pass := "pass"
		return &ibmsm.UsernamePasswordSecret{
			Name:     &name,
			ID:       &id,
			Username: &user,
			Password: &pass,
		}, nil, nil
	}
}

func (m *MockIBMSMClient) GetSecretVersion(getSecretOptions *ibmsm.GetSecretVersionOptions) (result ibmsm.SecretVersionIntf, response *core.DetailedResponse, err error) {
	m.GetSecretLock.Lock()
	m.GetSecretVersionCalledWith = getSecretOptions
	m.GetSecretVersionCallCount += 1
	m.GetSecretLock.Unlock()
	data := "dummy"
	id := "public_cert"
	yes := true
	return &ibmsm.PublicCertificateVersion{
		ID:               &id,
		PayloadAvailable: &yes,
		Certificate:      &data,
		PrivateKey:       &data,
		Intermediate:     &data,
	}, nil, nil
}

func TestIBMSecretsManagerGetSecrets(t *testing.T) {

	t.Run("Retrieves arbitrary secrets from a group", func(t *testing.T) {
		mock := MockIBMSMClient{}
		sm := backends.NewIBMSecretsManagerBackend(&mock)
		res, err := sm.GetSecrets("ibmcloud/arbitrary/secrets/groups/small-group", "", nil)
		if err != nil {
			t.FailNow()
		}

		// Properly calls ListSecrets the right number of times
		var offset int64 = 0
		expectedListArgs := &ibmsm.ListSecretsOptions{
			Groups: []string{"small-group"},
			Offset: &offset,
		}
		if !reflect.DeepEqual(mock.ListSecretsOptionCalledWith[0], expectedListArgs) {
			t.Errorf("expectedListArgs: %s, got: %s.", expectedListArgs.Groups, mock.ListSecretsOptionCalledWith[0].Groups)
		}
		if len(mock.ListSecretsOptionCalledWith) > 1 {
			t.Errorf("ListSecrets should be called %d times got %d", 1, len(mock.ListSecretsOptionCalledWith))
		}

		// Properly calls GetSecret
		id := "arbitrary"
		expectedGetArgs := &ibmsm.GetSecretOptions{
			ID: &id,
		}
		if mock.GetSecretCallCount != 1 {
			t.Errorf("GetSecret should be called %d times got %d", 1, mock.GetSecretCallCount)
		}
		if !reflect.DeepEqual(mock.GetSecretCalledWith, expectedGetArgs) {
			t.Errorf("Retrieved ID does not match expected, %s %s", *mock.GetSecretCalledWith.ID, *expectedGetArgs.ID)
		}

		// Correct data
		expected := map[string]interface{}{
			"my-secret": "password",
		}
		if !reflect.DeepEqual(res, expected) {
			t.Errorf("expected: %s, got: %s.", expected, res)
		}
	})

	t.Run("Paginates through groups with > IBMMaxPerPage secrets", func(t *testing.T) {
		mock := MockIBMSMClient{}
		sm := backends.NewIBMSecretsManagerBackend(&mock)

		res, err := sm.GetSecrets("ibmcloud/arbitrary/secrets/groups/big-group", "", nil)
		if err != nil {
			t.FailNow()
		}

		// Properly calls ListSecrets
		var offset int64 = 0
		var offset2 int64 = 200
		expectedListArgs := []*ibmsm.ListSecretsOptions{
			&ibmsm.ListSecretsOptions{
				Groups: []string{"big-group"},
				Offset: &offset,
			},
			&ibmsm.ListSecretsOptions{
				Groups: []string{"big-group"},
				Offset: &offset2,
			},
		}
		if len(mock.ListSecretsOptionCalledWith) != 2 {
			t.Fatalf("ListSecrets should be called %d times got %d", 2, len(mock.ListSecretsOptionCalledWith))
		}
		if !reflect.DeepEqual(mock.ListSecretsOptionCalledWith, expectedListArgs) {
			t.Errorf("ListSecrets was not called with the right arguments")
			t.Errorf("%d", *mock.ListSecretsOptionCalledWith[0].Offset)
			t.Errorf("%d", *mock.ListSecretsOptionCalledWith[1].Offset)
		}
		if len(res) != BIG_GROUP_LEN {
			t.Fatalf("GetSecrets did not retrieve all the secrets")
		}
	})

	t.Run("IBM SM GetIndividualSecret", func(t *testing.T) {
		mock := MockIBMSMClient{}
		sm := backends.NewIBMSecretsManagerBackend(&mock)

		secret, err := sm.GetIndividualSecret("ibmcloud/arbitrary/secrets/groups/small-group", "my-secret", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := "password"

		if !reflect.DeepEqual(expected, secret) {
			t.Errorf("expected: %s, got: %s", expected, secret)
		}
	})

	t.Run("Handles paths missing secret group and type", func(t *testing.T) {
		mock := MockIBMSMClient{}
		sm := backends.NewIBMSecretsManagerBackend(&mock)

		_, err := sm.GetSecrets("secret/data/my-secret", "", nil)
		if err == nil {
			t.FailNow()
		}
		expectedErr := "Path is not in the correct format"
		if !strings.Contains(err.Error(), expectedErr) {
			t.Fatalf("Expected error to have %s but said %s", expectedErr, err)
		}
	})

	t.Run("Retrieves payload of non-arbitrary, not-versioned secrets", func(t *testing.T) {
		mock := MockIBMSMClient{}
		sm := backends.NewIBMSecretsManagerBackend(&mock)

		res, err := sm.GetSecrets("ibmcloud/username_password/secrets/groups/small-group", "", nil)
		if err != nil {
			t.Fatalf("%s", err)
		}

		// Properly calls ListSecrets
		var offset int64 = 0
		expectedListArgs := &ibmsm.ListSecretsOptions{
			Groups: []string{"small-group"},
			Offset: &offset,
		}
		if !reflect.DeepEqual(mock.ListSecretsOptionCalledWith[0], expectedListArgs) {
			t.Errorf("expectedListArgs: %s, got: %s.", expectedListArgs.Groups, mock.ListSecretsOptionCalledWith[0].Groups)
		}
		if len(mock.ListSecretsOptionCalledWith) > 1 {
			t.Errorf("ListSecrets should be called %d times got %d", 1, len(mock.ListSecretsOptionCalledWith))
		}

		// Properly calls GetSecret
		id := "username_password"
		expectedGetArgs := &ibmsm.GetSecretOptions{
			ID: &id,
		}
		if !reflect.DeepEqual(mock.GetSecretCalledWith, expectedGetArgs) {
			t.Errorf("Retrieved ID and SecretType do not match expected")
		}

		// Correct data
		expected := map[string]interface{}{
			"my-secret": map[string]interface{}{
				"username": "user",
				"password": "pass",
			},
		}
		if !reflect.DeepEqual(res, expected) {
			t.Errorf("expected: %s, got: %s.", expected, res)
		}
	})

	t.Run("Retrieves payload of IAM credential secrets", func(t *testing.T) {
		mock := MockIBMSMClient{}
		sm := backends.NewIBMSecretsManagerBackend(&mock)

		res, err := sm.GetSecrets("ibmcloud/iam_credentials/secrets/groups/small-group", "", nil)
		if err != nil {
			t.Fatalf("%s", err)
		}

		// Properly calls ListSecrets
		var offset int64 = 0
		expectedListArgs := &ibmsm.ListSecretsOptions{
			Groups: []string{"small-group"},
			Offset: &offset,
		}
		if !reflect.DeepEqual(mock.ListSecretsOptionCalledWith[0], expectedListArgs) {
			t.Errorf("expectedListArgs: %s, got: %s.", expectedListArgs.Groups, mock.ListSecretsOptionCalledWith[0].Groups)
		}
		if len(mock.ListSecretsOptionCalledWith) > 1 {
			t.Errorf("ListSecrets should be called %d times got %d", 1, len(mock.ListSecretsOptionCalledWith))
		}

		// Properly calls GetSecret
		id := "iam_credentials"
		expectedGetArgs := &ibmsm.GetSecretOptions{
			ID: &id,
		}
		if !reflect.DeepEqual(mock.GetSecretCalledWith, expectedGetArgs) {
			t.Errorf("Retrieved ID and SecretType do not match expected")
		}

		// Correct data
		expected := map[string]interface{}{
			"my-secret": map[string]interface{}{
				"api_key": "password",
			},
		}
		if !reflect.DeepEqual(res, expected) {
			t.Errorf("expected: %s, got: %s.", expected, res)
		}
	})

	t.Run("Retrieves payload of KV secrets", func(t *testing.T) {
		mock := MockIBMSMClient{}
		sm := backends.NewIBMSecretsManagerBackend(&mock)

		res, err := sm.GetSecrets("ibmcloud/kv/secrets/groups/small-group", "", nil)
		if err != nil {
			t.Fatalf("%s", err)
		}

		// Properly calls ListSecrets
		var offset int64 = 0
		expectedListArgs := &ibmsm.ListSecretsOptions{
			Groups: []string{"small-group"},
			Offset: &offset,
		}
		if !reflect.DeepEqual(mock.ListSecretsOptionCalledWith[0], expectedListArgs) {
			t.Errorf("expectedListArgs: %s, got: %s.", expectedListArgs.Groups, mock.ListSecretsOptionCalledWith[0].Groups)
		}
		if len(mock.ListSecretsOptionCalledWith) > 1 {
			t.Errorf("ListSecrets should be called %d times got %d", 1, len(mock.ListSecretsOptionCalledWith))
		}

		// Properly calls GetSecret
		id := "kv"
		expectedGetArgs := &ibmsm.GetSecretOptions{
			ID: &id,
		}
		if !reflect.DeepEqual(mock.GetSecretCalledWith, expectedGetArgs) {
			t.Errorf("Retrieved ID and SecretType do not match expected")
		}

		// Correct data
		expected := map[string]interface{}{
			"my-secret": map[string]interface{}{
				"hello": "there",
			},
		}
		if !reflect.DeepEqual(res, expected) {
			t.Errorf("expected: %s, got: %s.", expected, res)
		}
	})

	t.Run("Properly retrieves versioned secrets", func(t *testing.T) {
		mock := MockIBMSMClient{}
		sm := backends.NewIBMSecretsManagerBackend(&mock)

		res, err := sm.GetSecrets("ibmcloud/public_cert/secrets/groups/small-group", "321", nil)
		if err != nil {
			t.Fatalf("%s", err)
		}

		// Properly calls ListSecrets
		var offset int64 = 0
		expectedListArgs := &ibmsm.ListSecretsOptions{
			Groups: []string{"small-group"},
			Offset: &offset,
		}
		if !reflect.DeepEqual(mock.ListSecretsOptionCalledWith[0], expectedListArgs) {
			t.Errorf("expectedListArgs: %s, got: %s.", expectedListArgs.Groups, mock.ListSecretsOptionCalledWith[0].Groups)
		}
		if len(mock.ListSecretsOptionCalledWith) > 1 {
			t.Errorf("ListSecrets should be called %d times got %d", 1, len(mock.ListSecretsOptionCalledWith))
		}

		// Properly calls GetSecretVersion
		id := "public_cert"
		version := "321"
		expectedGetArgs := &ibmsm.GetSecretVersionOptions{
			SecretID: &id,
			ID:       &version,
		}
		if !reflect.DeepEqual(mock.GetSecretVersionCalledWith, expectedGetArgs) {
			t.Errorf("Retrieved ID and SecretType do not match expected")
		}

		// Correct data
		expected := map[string]interface{}{
			"my-secret": map[string]interface{}{
				"certificate":  "dummy",
				"private_key":  "dummy",
				"intermediate": "dummy",
			},
		}
		if !reflect.DeepEqual(res, expected) {
			t.Errorf("expected: %s, got: %s.", expected, res)
		}
	})

	t.Run("Uses listSecrets cache only when listing from same group", func(t *testing.T) {
		mock := MockIBMSMClient{}
		sm := backends.NewIBMSecretsManagerBackend(&mock)

		// Call listSecrets API since cache is empty
		_, err := sm.GetIndividualSecret("ibmcloud/username_password/secrets/groups/small-group", "my-secret", "", nil)
		if err != nil {
			t.Fatalf("%s", err)
		}
		var offset int64 = 0
		expectedListArgs := &ibmsm.ListSecretsOptions{
			Groups: []string{"small-group"},
			Offset: &offset,
		}
		if !reflect.DeepEqual(mock.ListSecretsOptionCalledWith[0], expectedListArgs) {
			t.Errorf("expectedListArgs: %s, got: %s.", expectedListArgs.Groups, mock.ListSecretsOptionCalledWith[0].Groups)
		}
		if len(mock.ListSecretsOptionCalledWith) != 1 {
			t.Errorf("ListSecrets should be called %d times got %d", 1, len(mock.ListSecretsOptionCalledWith))
		}

		// Serve from cache since populated for groupId small-group
		_, err = sm.GetIndividualSecret("ibmcloud/arbitrary/secrets/groups/small-group", "my-secret", "", nil)
		if err != nil {
			t.Fatalf("%s", err)
		}
		if len(mock.ListSecretsOptionCalledWith) != 1 {
			t.Errorf("ListSecrets should be called %d times got %d", 1, len(mock.ListSecretsOptionCalledWith))
		}

		// Call API again since no cached data for groupId empty-group
		_, err = sm.GetIndividualSecret("ibmcloud/arbitrary/secrets/groups/empty-group", "my-secret", "", nil)
		if err != nil {
			t.Fatalf("%s", err)
		}
		if len(mock.ListSecretsOptionCalledWith) != 2 {
			t.Errorf("ListSecrets should be called %d times got %d", 2, len(mock.ListSecretsOptionCalledWith))
		}
	})

	t.Run("Uses getSecrets cache only when reading from same group, same type", func(t *testing.T) {
		mock := MockIBMSMClient{}
		sm := backends.NewIBMSecretsManagerBackend(&mock)

		// Call API since no cached data for small-group
		_, err := sm.GetSecrets("ibmcloud/arbitrary/secrets/groups/small-group", "", nil)
		if err != nil {
			t.Fatalf("%s", err)
		}
		if mock.GetSecretCallCount != 1 {
			t.Errorf("GetSecret should be called %d times got %d", 1, mock.GetSecretCallCount)
		}

		// Serve from cache since all secrets retrieved for small-group
		_, err = sm.GetIndividualSecret("ibmcloud/arbitrary/secrets/groups/small-group", "my-secret", "", nil)
		if err != nil {
			t.Fatalf("%s", err)
		}
		if mock.GetSecretCallCount != 1 {
			t.Errorf("GetSecret should be called %d times got %d", 1, mock.GetSecretCallCount)
		}

		// Serve from cache since all secrets retrieved for small-group
		_, err = sm.GetSecrets("ibmcloud/arbitrary/secrets/groups/small-group", "", nil)
		if err != nil {
			t.Fatalf("%s", err)
		}
		if mock.GetSecretCallCount != 1 {
			t.Errorf("GetSecret should be called %d times got %d", 1, mock.GetSecretCallCount)
		}

		// Call API again since no cached data for username_password secrets of group small-group
		_, err = sm.GetIndividualSecret("ibmcloud/username_password/secrets/groups/small-group", "my-secret", "", nil)
		if err != nil {
			t.Fatalf("%s", err)
		}
		if mock.GetSecretCallCount != 2 {
			t.Errorf("GetSecret should be called %d times got %d", 2, mock.GetSecretCallCount)
		}

		// Call API again since no cached data for arbitrary secrets of group big-group
		_, err = sm.GetIndividualSecret("ibmcloud/arbitrary/secrets/groups/big-group", "my-secret-1", "", nil)
		if err != nil {
			t.Fatalf("%s", err)
		}
		if mock.GetSecretCallCount != 3 {
			t.Errorf("GetSecret should be called %d times got %d", 3, mock.GetSecretCallCount)
		}
	})

	t.Run("Bypasses getSecrets cache when reading specific version of secret", func(t *testing.T) {
		mock := MockIBMSMClient{}
		sm := backends.NewIBMSecretsManagerBackend(&mock)

		// Call API since no cached data for small-group v2
		_, err := sm.GetSecrets("ibmcloud/public_cert/secrets/groups/small-group", "v2", nil)
		if err != nil {
			t.Fatalf("%s", err)
		}
		if mock.GetSecretVersionCallCount != 1 {
			t.Errorf("GetSecret should be called %d times got %d", 1, mock.GetSecretVersionCallCount)
		}

		// Bypass cache again b/c specific version requested
		_, err = sm.GetSecrets("ibmcloud/public_cert/secrets/groups/small-group", "v2", nil)
		if err != nil {
			t.Fatalf("%s", err)
		}
		if mock.GetSecretVersionCallCount != 2 {
			t.Errorf("GetSecret should be called %d times got %d", 2, mock.GetSecretVersionCallCount)
		}

		// Bypass cache again b/c specific version requested
		_, err = sm.GetIndividualSecret("ibmcloud/public_cert/secrets/groups/small-group", "my-secret", "v2", nil)
		if err != nil {
			t.Fatalf("%s", err)
		}
		if mock.GetSecretVersionCallCount != 3 {
			t.Errorf("GetIndividualSecret should be called %d times got %d", 3, mock.GetSecretCallCount)
		}
	})
}
