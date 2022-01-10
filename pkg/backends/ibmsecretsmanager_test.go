package backends_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/IBM/go-sdk-core/v5/core"
	ibmsm "github.com/IBM/secrets-manager-go-sdk/secretsmanagerv1"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/backends"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/types"
)

type MockIBMSMClient struct {
	ListAllSecretsOptionCalledWith []*ibmsm.ListAllSecretsOptions
	GetSecretCalledWith            *ibmsm.GetSecretOptions
	GetSecretCallCount             int
	GetSecretVersionCalledWith     *ibmsm.GetSecretVersionOptions
}

var BIG_GROUP_LEN int = types.IBMMaxPerPage + 1

// This is used to take deep copies of struct fields passed as pointers in ListAllSecretsOptions
// so we can make assertions about the values later
func deepCopy(listAllSecretsOptions *ibmsm.ListAllSecretsOptions) *ibmsm.ListAllSecretsOptions {
	var offset int64 = *listAllSecretsOptions.Offset
	return &ibmsm.ListAllSecretsOptions{
		Groups: listAllSecretsOptions.Groups,
		Offset: &offset,
	}
}

func (m *MockIBMSMClient) ListAllSecrets(listAllSecretsOptions *ibmsm.ListAllSecretsOptions) (result *ibmsm.ListSecrets, response *core.DetailedResponse, err error) {
	m.ListAllSecretsOptionCalledWith = append(m.ListAllSecretsOptionCalledWith, deepCopy(listAllSecretsOptions))

	// A big secret group
	bigGroup := "big-group"
	stype := "arbitrary"
	bigGroupSecrets := make([]ibmsm.SecretResourceIntf, BIG_GROUP_LEN)
	for id := 0; id < BIG_GROUP_LEN; id += 1 {
		name := fmt.Sprintf("my-secret-%d", id)
		bigGroupSecrets[id] = &ibmsm.SecretResource{
			Name:          &name,
			SecretType:    &stype,
			SecretGroupID: &bigGroup,
			ID:            &name,
		}
	}

	// A small secret  group
	smallGroup := "small-group"
	name := "my-secret"
	id := "123"
	otype := "username_password"
	ctype := "public_cert"
	itype := "iam_credentials"
	smallGroupSecrets := []ibmsm.SecretResourceIntf{
		&ibmsm.SecretResource{
			Name:          &name,
			SecretType:    &stype,
			SecretGroupID: &smallGroup,
			ID:            &id,
		},
		&ibmsm.SecretResource{
			Name:          &name,
			SecretType:    &otype,
			SecretGroupID: &smallGroup,
			ID:            &id,
		},
		&ibmsm.SecretResource{
			Name:          &name,
			SecretType:    &ctype,
			SecretGroupID: &smallGroup,
			ID:            &id,
		},
		&ibmsm.SecretResource{
			Name:          &name,
			SecretType:    &itype,
			SecretGroupID: &smallGroup,
			ID:            &id,
		},
	}

	// Empty secret group
	emptyGroup := "empty-group"
	emptyGroupSecrets := []ibmsm.SecretResourceIntf{}

	if listAllSecretsOptions.Groups[0] == bigGroup {
		// Emulate a 2-page paginated response
		offset := int(*listAllSecretsOptions.Offset)

		end := offset + types.IBMMaxPerPage
		if end > BIG_GROUP_LEN {
			end = BIG_GROUP_LEN
		}

		return &ibmsm.ListSecrets{
			Resources: bigGroupSecrets[offset:end],
		}, nil, nil
	} else if listAllSecretsOptions.Groups[0] == smallGroup {
		return &ibmsm.ListSecrets{
			Resources: smallGroupSecrets,
		}, nil, nil
	} else if listAllSecretsOptions.Groups[0] == emptyGroup {
		return &ibmsm.ListSecrets{
			Resources: emptyGroupSecrets,
		}, nil, nil
	} else {
		return nil, nil, fmt.Errorf("No such group %s", listAllSecretsOptions.Groups[0])
	}
}

func (m *MockIBMSMClient) GetSecret(getSecretOptions *ibmsm.GetSecretOptions) (result *ibmsm.GetSecret, response *core.DetailedResponse, err error) {
	m.GetSecretCalledWith = getSecretOptions
	m.GetSecretCallCount += 1

	if *getSecretOptions.SecretType == "arbitrary" {
		name := "my-secret"
		id := "123"
		payload := "password"
		return &ibmsm.GetSecret{
			Resources: []ibmsm.SecretResourceIntf{
				&ibmsm.SecretResource{
					Name: &name,
					ID:   &id,
					SecretData: map[string]interface{}{
						"payload": payload,
					},
				},
			},
		}, nil, nil
	} else if *getSecretOptions.SecretType == "iam_credentials" {
		name := "my-secret"
		id := "123"
		payload := "password"
		return &ibmsm.GetSecret{
			Resources: []ibmsm.SecretResourceIntf{
				&ibmsm.SecretResource{
					Name:   &name,
					ID:     &id,
					APIKey: &payload,
				},
			},
		}, nil, nil
	} else {
		name := "my-secret"
		id := "123"
		return &ibmsm.GetSecret{
			Resources: []ibmsm.SecretResourceIntf{
				&ibmsm.SecretResource{
					Name: &name,
					ID:   &id,
					SecretData: map[string]interface{}{
						"username": "user",
						"password": "pass",
					},
				},
			},
		}, nil, nil
	}
}

func (m *MockIBMSMClient) GetSecretVersion(getSecretOptions *ibmsm.GetSecretVersionOptions) (result *ibmsm.GetSecretVersion, response *core.DetailedResponse, err error) {
	m.GetSecretVersionCalledWith = getSecretOptions
	data := "dummy"
	id := "123"
	return &ibmsm.GetSecretVersion{
		Resources: []ibmsm.SecretVersionIntf{
			&ibmsm.SecretVersion{
				ID: &id,
				SecretData: map[string]interface{}{
					"certificate": data,
				},
			},
		},
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
		expectedListArgs := &ibmsm.ListAllSecretsOptions{
			Groups: []string{"small-group"},
			Offset: &offset,
		}
		if !reflect.DeepEqual(mock.ListAllSecretsOptionCalledWith[0], expectedListArgs) {
			t.Errorf("expectedListArgs: %s, got: %s.", expectedListArgs.Groups, mock.ListAllSecretsOptionCalledWith[0].Groups)
		}
		if len(mock.ListAllSecretsOptionCalledWith) > 1 {
			t.Errorf("ListAllSecrets should be called %d times got %d", 1, len(mock.ListAllSecretsOptionCalledWith))
		}

		// Properly calls GetSecret
		id := "123"
		stype := "arbitrary"
		expectedGetArgs := &ibmsm.GetSecretOptions{
			ID:         &id,
			SecretType: &stype,
		}
		if !reflect.DeepEqual(mock.GetSecretCalledWith, expectedGetArgs) {
			t.Errorf("Retrieved ID and SecretType do not match expected")
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
		expectedListArgs := []*ibmsm.ListAllSecretsOptions{
			&ibmsm.ListAllSecretsOptions{
				Groups: []string{"big-group"},
				Offset: &offset,
			},
			&ibmsm.ListAllSecretsOptions{
				Groups: []string{"big-group"},
				Offset: &offset2,
			},
		}
		if len(mock.ListAllSecretsOptionCalledWith) != 2 {
			t.Fatalf("ListAllSecrets should be called %d times got %d", 2, len(mock.ListAllSecretsOptionCalledWith))
		}
		if !reflect.DeepEqual(mock.ListAllSecretsOptionCalledWith, expectedListArgs) {
			t.Errorf("ListAllSecrets was not called with the right arguments")
			t.Errorf("%d", *mock.ListAllSecretsOptionCalledWith[0].Offset)
			t.Errorf("%d", *mock.ListAllSecretsOptionCalledWith[1].Offset)
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
		expectedListArgs := &ibmsm.ListAllSecretsOptions{
			Groups: []string{"small-group"},
			Offset: &offset,
		}
		if !reflect.DeepEqual(mock.ListAllSecretsOptionCalledWith[0], expectedListArgs) {
			t.Errorf("expectedListArgs: %s, got: %s.", expectedListArgs.Groups, mock.ListAllSecretsOptionCalledWith[0].Groups)
		}
		if len(mock.ListAllSecretsOptionCalledWith) > 1 {
			t.Errorf("ListAllSecrets should be called %d times got %d", 1, len(mock.ListAllSecretsOptionCalledWith))
		}

		// Properly calls GetSecret
		id := "123"
		stype := "username_password"
		expectedGetArgs := &ibmsm.GetSecretOptions{
			ID:         &id,
			SecretType: &stype,
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
		expectedListArgs := &ibmsm.ListAllSecretsOptions{
			Groups: []string{"small-group"},
			Offset: &offset,
		}
		if !reflect.DeepEqual(mock.ListAllSecretsOptionCalledWith[0], expectedListArgs) {
			t.Errorf("expectedListArgs: %s, got: %s.", expectedListArgs.Groups, mock.ListAllSecretsOptionCalledWith[0].Groups)
		}
		if len(mock.ListAllSecretsOptionCalledWith) > 1 {
			t.Errorf("ListAllSecrets should be called %d times got %d", 1, len(mock.ListAllSecretsOptionCalledWith))
		}

		// Properly calls GetSecret
		id := "123"
		stype := "iam_credentials"
		expectedGetArgs := &ibmsm.GetSecretOptions{
			ID:         &id,
			SecretType: &stype,
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

	t.Run("Properly retrieves versioned secrets", func(t *testing.T) {
		mock := MockIBMSMClient{}
		sm := backends.NewIBMSecretsManagerBackend(&mock)

		res, err := sm.GetSecrets("ibmcloud/public_cert/secrets/groups/small-group", "321", nil)
		if err != nil {
			t.Fatalf("%s", err)
		}

		// Properly calls ListSecrets
		var offset int64 = 0
		expectedListArgs := &ibmsm.ListAllSecretsOptions{
			Groups: []string{"small-group"},
			Offset: &offset,
		}
		if !reflect.DeepEqual(mock.ListAllSecretsOptionCalledWith[0], expectedListArgs) {
			t.Errorf("expectedListArgs: %s, got: %s.", expectedListArgs.Groups, mock.ListAllSecretsOptionCalledWith[0].Groups)
		}
		if len(mock.ListAllSecretsOptionCalledWith) > 1 {
			t.Errorf("ListAllSecrets should be called %d times got %d", 1, len(mock.ListAllSecretsOptionCalledWith))
		}

		// Properly calls GetSecretVersion
		id := "123"
		stype := "public_cert"
		version := "321"
		expectedGetArgs := &ibmsm.GetSecretVersionOptions{
			ID:         &id,
			SecretType: &stype,
			VersionID:  &version,
		}
		if !reflect.DeepEqual(mock.GetSecretVersionCalledWith, expectedGetArgs) {
			t.Errorf("Retrieved ID and SecretType do not match expected")
		}

		// Correct data
		expected := map[string]interface{}{
			"my-secret": map[string]interface{}{
				"certificate": "dummy",
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
		expectedListArgs := &ibmsm.ListAllSecretsOptions{
			Groups: []string{"small-group"},
			Offset: &offset,
		}
		if !reflect.DeepEqual(mock.ListAllSecretsOptionCalledWith[0], expectedListArgs) {
			t.Errorf("expectedListArgs: %s, got: %s.", expectedListArgs.Groups, mock.ListAllSecretsOptionCalledWith[0].Groups)
		}
		if len(mock.ListAllSecretsOptionCalledWith) != 1 {
			t.Errorf("ListAllSecrets should be called %d times got %d", 1, len(mock.ListAllSecretsOptionCalledWith))
		}

		// Serve from cache since populated for groupId small-group
		_, err = sm.GetIndividualSecret("ibmcloud/arbitrary/secrets/groups/small-group", "my-secret", "", nil)
		if err != nil {
			t.Fatalf("%s", err)
		}
		if len(mock.ListAllSecretsOptionCalledWith) != 1 {
			t.Errorf("ListAllSecrets should be called %d times got %d", 1, len(mock.ListAllSecretsOptionCalledWith))
		}

		// Call API again since no cached data for groupId empty-group
		_, err = sm.GetIndividualSecret("ibmcloud/arbitrary/secrets/groups/empty-group", "my-secret", "", nil)
		if err != nil {
			t.Fatalf("%s", err)
		}
		if len(mock.ListAllSecretsOptionCalledWith) != 2 {
			t.Errorf("ListAllSecrets should be called %d times got %d", 2, len(mock.ListAllSecretsOptionCalledWith))
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
		_, err := sm.GetSecrets("ibmcloud/arbitrary/secrets/groups/small-group", "v2", nil)
		if err != nil {
			t.Fatalf("%s", err)
		}
		if mock.GetSecretCallCount != 1 {
			t.Errorf("GetSecret should be called %d times got %d", 1, mock.GetSecretCallCount)
		}

		// Bypass cache again b/c specific version requested
		_, err = sm.GetSecrets("ibmcloud/arbitrary/secrets/groups/small-group", "v2", nil)
		if err != nil {
			t.Fatalf("%s", err)
		}
		if mock.GetSecretCallCount != 2 {
			t.Errorf("GetSecret should be called %d times got %d", 2, mock.GetSecretCallCount)
		}

		// Bypass cache again b/c specific version requested
		_, err = sm.GetIndividualSecret("ibmcloud/arbitrary/secrets/groups/small-group", "my-secret", "v2", nil)
		if err != nil {
			t.Fatalf("%s", err)
		}
		if mock.GetSecretCallCount != 3 {
			t.Errorf("GetIndividualSecret should be called %d times got %d", 3, mock.GetSecretCallCount)
		}
	})
}
