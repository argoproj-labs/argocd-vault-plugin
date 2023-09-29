package backends_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/backends"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type mockSecretsManagerClient struct {
	backends.AWSSecretsManagerIface
}

func (m *mockSecretsManagerClient) GetSecretValue(ctx context.Context, input *secretsmanager.GetSecretValueInput, options ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	data := &secretsmanager.GetSecretValueOutput{}

	switch *input.SecretId {
	case "test":
		if input.VersionId == nil {
			string := "{\"test-secret\":\"current-value\"}"
			data.SecretString = &string
		} else {
			string := "{\"test-secret\":\"previous-value\"}"
			data.SecretString = &string
		}
	case "test-binary":
		data.SecretBinary = []byte("binary-data")
	}

	return data, nil
}

func TestAWSSecretManagerGetSecrets(t *testing.T) {
	sm := backends.NewAWSSecretsManagerBackend(&mockSecretsManagerClient{})

	t.Run("Get secrets", func(t *testing.T) {
		data, err := sm.GetSecrets("test", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"test-secret": "current-value",
		}

		if !reflect.DeepEqual(expected, data) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}
	})

	t.Run("AWS GetIndividualSecret", func(t *testing.T) {
		secret, err := sm.GetIndividualSecret("test", "test-secret", "previous", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := "previous-value"

		if !reflect.DeepEqual(expected, secret) {
			t.Errorf("expected: %s, got: %s.", expected, secret)
		}
	})

	t.Run("Get secrets at specific version", func(t *testing.T) {
		data, err := sm.GetSecrets("test", "123", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"test-secret": "previous-value",
		}

		if !reflect.DeepEqual(expected, data) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}
	})

	t.Run("Get binary secret", func(t *testing.T) {
		data, err := sm.GetSecrets("test-binary", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"SecretBinary": []byte("binary-data"),
		}

		if !reflect.DeepEqual(expected, data) {
			t.Errorf("expected: %v, got: %v.", expected, data)
		}
	})
}

func TestAWSSecretManagerEmptyIfNoSecret(t *testing.T) {
	sm := backends.NewAWSSecretsManagerBackend(&mockSecretsManagerClient{})

	_, err := sm.GetSecrets("empty", "", map[string]string{})
	if err == nil {
		t.Fatalf("expected an error but got nil")
	}

	if err.Error() != "Could not find secret empty" {
		t.Errorf("expected error: %s, got: %s.", "Could not find secret empty", err.Error())
	}
}
