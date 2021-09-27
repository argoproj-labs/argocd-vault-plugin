package backends_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/IBM/argocd-vault-plugin/pkg/backends"
	"github.com/IBM/argocd-vault-plugin/pkg/types"
	"github.com/IBM/go-sdk-core/v5/core"
	ibmsm "github.com/IBM/secrets-manager-go-sdk/secretsmanagerv1"
)

type MockIBMSMClient struct {
	ListAllSecretsOptionCalledWith []*ibmsm.ListAllSecretsOptions
	GetSecretCalledWith            *ibmsm.GetSecretOptions
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
	stype := "arbitrary"
	bigGroupSecrets := make([]ibmsm.SecretResourceIntf, BIG_GROUP_LEN)
	for id := 0; id < BIG_GROUP_LEN; id += 1 {
		name := fmt.Sprintf("my-secret-%d", id)
		bigGroupSecrets[id] = &ibmsm.SecretResource{
			Name:       &name,
			SecretType: &stype,
			ID:         &name,
		}
	}

	// A small secret  group
	name := "my-secret"
	id := "123"
	otype := "username_password"
	smallGroupSecrets := []ibmsm.SecretResourceIntf{
		&ibmsm.SecretResource{
			Name:       &name,
			SecretType: &stype,
			ID:         &id,
		},
		&ibmsm.SecretResource{
			Name:       &name,
			SecretType: &otype,
			ID:         &id,
		},
	}

	if listAllSecretsOptions.Groups[0] == "big-group" {
		// Emulate a 2-page paginated response
		offset := int(*listAllSecretsOptions.Offset)

		end := offset + types.IBMMaxPerPage
		if end > BIG_GROUP_LEN {
			end = BIG_GROUP_LEN
		}

		return &ibmsm.ListSecrets{
			Resources: bigGroupSecrets[offset:end],
		}, nil, nil
	} else {
		return &ibmsm.ListSecrets{
			Resources: smallGroupSecrets,
		}, nil, nil
	}
}

func (m *MockIBMSMClient) GetSecret(getSecretOptions *ibmsm.GetSecretOptions) (result *ibmsm.GetSecret, response *core.DetailedResponse, err error) {
	m.GetSecretCalledWith = getSecretOptions

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

func TestIBMSecretsManagerGetSecrets(t *testing.T) {

	t.Run("Retrieves arbitrary secrets from a group", func(t *testing.T) {
		mock := MockIBMSMClient{}
		sm := backends.NewIBMSecretsManagerBackend(&mock)
		res, err := sm.GetSecrets("ibmcloud/arbitrary/secrets/groups/test-group-id", "", nil)
		if err != nil {
			t.FailNow()
		}

		// Properly calls ListSecrets the right number of times
		var offset int64 = 0
		expectedListArgs := &ibmsm.ListAllSecretsOptions{
			Groups: []string{"test-group-id"},
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

	t.Run("Helpful message for secrets that are not `arbitrary` type", func(t *testing.T) {
		mock := MockIBMSMClient{}
		sm := backends.NewIBMSecretsManagerBackend(&mock)

		m, err := sm.GetSecrets("ibmcloud/username_password/secrets/groups/test-group-id", "", nil)
		if err == nil {
			t.Fatalf("%s", m)
		}

		// Properly calls ListSecrets
		var offset int64 = 0
		expectedListArgs := &ibmsm.ListAllSecretsOptions{
			Groups: []string{"test-group-id"},
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

		// Errors bc the secret type is not `arbitrary` and therefore has no `payload`
		expectedErr := "No `payload` key present for secret"
		if !strings.Contains(err.Error(), expectedErr) {
			t.Fatalf("Expected error to have %s but said %s", expectedErr, err)
		}
	})
}
