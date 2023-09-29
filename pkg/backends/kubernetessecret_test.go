package backends

import (
	"reflect"
	"testing"
)

func newMockK8sClient(vals map[string]map[string]string, err error) *mockK8sClient {
	encoded := make(map[string]map[string][]byte)
	for path, secrets := range vals {
		encoded[path] = make(map[string][]byte)
		for key, value := range secrets {
			encoded[path][key] = []byte(value)
		}
	}
	return &mockK8sClient{
		responses_by_path: encoded,
		err:               err,
	}
}

type mockK8sClient struct {
	responses_by_path map[string]map[string][]byte
	err               error
}

func (m *mockK8sClient) ReadSecretData(path string) (map[string][]byte, error) {
	return m.responses_by_path[path], m.err
}

func TestKubernetesSecretGetSecrets(t *testing.T) {
	sm := NewKubernetesSecret()
	sm.client = newMockK8sClient(map[string]map[string]string{
		"secret1": {
			"test-secret": "current-value",
			"test2":       "bar",
		},
		"secret2": {
			"key": "foz",
		},
	}, nil)

	t.Run("Get secrets from first path", func(t *testing.T) {
		data, err := sm.GetSecrets("secret1", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"test-secret": "current-value",
			"test2":       "bar",
		}

		if !reflect.DeepEqual(expected, data) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}
	})

	t.Run("GetIndividualSecret from first path", func(t *testing.T) {
		secret, err := sm.GetIndividualSecret("secret1", "test2", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := "bar"

		if !reflect.DeepEqual(expected, secret) {
			t.Errorf("expected: %s, got: %s.", expected, secret)
		}
	})

	t.Run("Get secrets from secret from second path", func(t *testing.T) {
		data, err := sm.GetSecrets("secret2", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"key": "foz",
		}

		if !reflect.DeepEqual(expected, data) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}
	})

	t.Run("GetIndividualSecret from inline path secret", func(t *testing.T) {
		secret, err := sm.GetIndividualSecret("secret2", "key", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := "foz"

		if !reflect.DeepEqual(expected, secret) {
			t.Errorf("expected: %s, got: %s.", expected, secret)
		}
	})

	t.Run("GetIndividualSecretNotFound", func(t *testing.T) {
		secret, err := sm.GetIndividualSecret("test", "22test2", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		if secret != nil {
			t.Errorf("expected: %s, got: %s.", "nil", secret)
		}
	})
}
