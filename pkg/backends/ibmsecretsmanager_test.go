package backends_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/IBM/argocd-vault-plugin/pkg/backends"
	"github.com/IBM/go-sdk-core/v5/core"
	ibmsm "github.com/IBM/secrets-manager-go-sdk/secretsmanagerv1"
)

type MockIBMSMClient struct {
	ListAllSecretsOptionCalledWith *ibmsm.ListAllSecretsOptions
	GetSecretCalledWith            *ibmsm.GetSecretOptions
}

func (m *MockIBMSMClient) ListAllSecrets(listAllSecretsOptions *ibmsm.ListAllSecretsOptions) (result *ibmsm.ListSecrets, response *core.DetailedResponse, err error) {
	m.ListAllSecretsOptionCalledWith = listAllSecretsOptions
	name := "my-secret"
	id := "123"
	stype := "arbitrary"
	otype := "username_password"
	return &ibmsm.ListSecrets{
		Resources: []ibmsm.SecretResourceIntf{
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
		},
	}, nil, nil
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

	mock := MockIBMSMClient{}
	sm := backends.NewIBMSecretsManagerBackend(&mock)

	t.Run("Retrieves arbitrary secrets from a group", func(t *testing.T) {
		res, err := sm.GetSecrets("ibmcloud/arbitrary/secrets/groups/test-group-id", "", nil)
		if err != nil {
			t.FailNow()
		}

		// Properly calls ListSecrets
		expectedListArgs := &ibmsm.ListAllSecretsOptions{
			Groups: []string{"test-group-id"},
		}
		if !reflect.DeepEqual(mock.ListAllSecretsOptionCalledWith, expectedListArgs) {
			t.Errorf("expectedListArgs: %s, got: %s.", expectedListArgs.Groups, mock.ListAllSecretsOptionCalledWith.Groups)
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

	t.Run("Handles paths missing secret group and type", func(t *testing.T) {
		_, err := sm.GetSecretsVersioned("secret/data/my-secret", "", nil)
		if err == nil {
			t.FailNow()
		}
		expectedErr := "Path is not in the correct format"
		if !strings.Contains(err.Error(), expectedErr) {
			t.Fatalf("Expected error to have %s but said %s", expectedErr, err)
		}
	})

	t.Run("Helpful message for secrets that are not `arbitrary` type", func(t *testing.T) {
		m, err := sm.GetSecrets("ibmcloud/username_password/secrets/groups/test-group-id", "", nil)
		if err == nil {
			t.Fatalf("%s", m)
		}

		// Properly calls ListSecrets
		expectedListArgs := &ibmsm.ListAllSecretsOptions{
			Groups: []string{"test-group-id"},
		}
		if !reflect.DeepEqual(mock.ListAllSecretsOptionCalledWith, expectedListArgs) {
			t.Errorf("expectedListArgs: %s, got: %s.", expectedListArgs.Groups, mock.ListAllSecretsOptionCalledWith.Groups)
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
