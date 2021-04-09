package backends_test

import (
	"reflect"
	"testing"

	"github.com/IBM/argocd-vault-plugin/pkg/backends"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
)

type mockSecretsManagerClient struct {
	secretsmanageriface.SecretsManagerAPI
}

func (m *mockSecretsManagerClient) GetSecretValue(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
	data := &secretsmanager.GetSecretValueOutput{}

	switch *input.SecretId {
	case "test":
		string := "{\"test-secret\":\"some-value\"}"
		data.SecretString = &string
	}

	return data, nil
}

func TestAWSSecretManagerGetSecrets(t *testing.T) {
	sm := backends.NewAWSSecretsManagerBackend(&mockSecretsManagerClient{})

	data, err := sm.GetSecrets("test", map[string]string{})
	if err != nil {
		t.Fatalf("expected 0 errors but got: %s", err)
	}

	expected := map[string]interface{}{
		"test-secret": "some-value",
	}

	if !reflect.DeepEqual(expected, data) {
		t.Errorf("expected: %s, got: %s.", expected, data)
	}
}

func TestAWSSecretManagerEmptyIfNoSecret(t *testing.T) {
	sm := backends.NewAWSSecretsManagerBackend(&mockSecretsManagerClient{})

	_, err := sm.GetSecrets("empty", map[string]string{})
	if err == nil {
		t.Fatalf("expected an error but got nil")
	}

	if err.Error() != "Could not find secret empty" {
		t.Errorf("expected error: %s, got: %s.", "Could not find secret empty", err.Error())
	}
}
