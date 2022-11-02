package backends_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	yaml "sigs.k8s.io/yaml"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/backends"
)

func TestLocalSecretManagerGetSecretsJson(t *testing.T) {

	sm := backends.NewLocalSecretManagerBackend(func(path string, format string) ([]byte, error) {
		data := map[string]interface{}{
			"test-secret": "current-value",
		}
		return json.Marshal(data)
	})

	t.Run("Get secrets", func(t *testing.T) {

		data, err := sm.GetSecrets("example.json", "", map[string]string{})
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

	t.Run("GetIndividualSecret", func(t *testing.T) {
		secret, err := sm.GetIndividualSecret("example.yaml", "test-secret", "previous", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := "current-value"

		if !reflect.DeepEqual(expected, secret) {
			t.Errorf("expected: %s, got: %s.", expected, secret)
		}
	})
}

func TestLocalSecretManagerGetSecretsYaml(t *testing.T) {

	sm := backends.NewLocalSecretManagerBackend(func(path string, format string) ([]byte, error) {
		data := map[string]interface{}{
			"test-secret": "current-value",
		}
		return yaml.Marshal(data)
	})

	t.Run("Get secrets", func(t *testing.T) {
		data, err := sm.GetSecrets("example.yaml", "", map[string]string{})
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
}

func TestLocalSecretManagerPathNotFound(t *testing.T) {
	sm := backends.NewLocalSecretManagerBackend(func(path string, format string) ([]byte, error) {
		return nil, fmt.Errorf("File %s not found", path)
	})

	_, err := sm.GetSecrets("empty", "", map[string]string{})
	if err == nil {
		t.Fatalf("expected an error but got nil")
	}

	expectedError := "Could not find file empty - File empty not found"
	if err.Error() != expectedError {
		t.Errorf("expected error: %s, got: %s.", expectedError, err.Error())
	}
}
