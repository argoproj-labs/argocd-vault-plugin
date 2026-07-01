package backends

import (
	"fmt"
	"reflect"
	"testing"

	infisical "github.com/infisical/go-sdk"
)

type mockInfisicalClient struct {
	secrets map[string]map[string]string
	err     error
}

func (m *mockInfisicalClient) List(options infisical.ListSecretsOptions) ([]infisical.Secret, error) {
	if m.err != nil {
		return nil, m.err
	}
	secretMap, ok := m.secrets[options.SecretPath]
	if !ok {
		return nil, nil
	}
	var result []infisical.Secret
	for k, v := range secretMap {
		result = append(result, infisical.Secret{
			SecretKey:   k,
			SecretValue: v,
		})
	}
	return result, nil
}

func (m *mockInfisicalClient) Retrieve(options infisical.RetrieveSecretOptions) (infisical.Secret, error) {
	if m.err != nil {
		return infisical.Secret{}, m.err
	}
	secretMap, ok := m.secrets[options.SecretPath]
	if !ok {
		return infisical.Secret{}, fmt.Errorf("path not found: %s", options.SecretPath)
	}
	value, ok := secretMap[options.SecretKey]
	if !ok {
		return infisical.Secret{}, fmt.Errorf("secret not found: %s", options.SecretKey)
	}
	return infisical.Secret{
		SecretKey:   options.SecretKey,
		SecretValue: value,
	}, nil
}

func newMockInfisicalClient(secrets map[string]map[string]string, err error) *mockInfisicalClient {
	return &mockInfisicalClient{
		secrets: secrets,
		err:     err,
	}
}

func TestInfisicalGetSecrets(t *testing.T) {
	mock := newMockInfisicalClient(map[string]map[string]string{
		"/": {
			"DB_HOST":     "localhost",
			"DB_PASSWORD": "secret123",
		},
		"/api": {
			"API_KEY": "key-value",
		},
	}, nil)

	sm := NewInfisicalBackend(mock, "my-project", "dev")

	t.Run("GetSecrets from root path", func(t *testing.T) {
		data, err := sm.GetSecrets("/", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"DB_HOST":     "localhost",
			"DB_PASSWORD": "secret123",
		}

		if !reflect.DeepEqual(expected, data) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}
	})

	t.Run("GetSecrets from subfolder path", func(t *testing.T) {
		data, err := sm.GetSecrets("/api", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"API_KEY": "key-value",
		}

		if !reflect.DeepEqual(expected, data) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}
	})

	t.Run("GetSecrets with empty path defaults to root", func(t *testing.T) {
		data, err := sm.GetSecrets("", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"DB_HOST":     "localhost",
			"DB_PASSWORD": "secret123",
		}

		if !reflect.DeepEqual(expected, data) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}
	})

	t.Run("GetSecrets returns empty map for unknown path", func(t *testing.T) {
		data, err := sm.GetSecrets("/unknown", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		if len(data) != 0 {
			t.Errorf("expected empty map, got: %s.", data)
		}
	})
}

func TestInfisicalGetSecretsError(t *testing.T) {
	mock := newMockInfisicalClient(nil, fmt.Errorf("api error"))
	sm := NewInfisicalBackend(mock, "my-project", "dev")

	_, err := sm.GetSecrets("/", "", map[string]string{})
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
}

func TestInfisicalGetIndividualSecret(t *testing.T) {
	mock := newMockInfisicalClient(map[string]map[string]string{
		"/": {
			"DB_HOST":     "localhost",
			"DB_PASSWORD": "secret123",
		},
		"/api": {
			"API_KEY": "key-value",
		},
	}, nil)

	sm := NewInfisicalBackend(mock, "my-project", "dev")

	t.Run("GetIndividualSecret from root", func(t *testing.T) {
		secret, err := sm.GetIndividualSecret("/", "DB_PASSWORD", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		if secret != "secret123" {
			t.Errorf("expected: secret123, got: %s.", secret)
		}
	})

	t.Run("GetIndividualSecret from subfolder", func(t *testing.T) {
		secret, err := sm.GetIndividualSecret("/api", "API_KEY", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		if secret != "key-value" {
			t.Errorf("expected: key-value, got: %s.", secret)
		}
	})

	t.Run("GetIndividualSecret with empty path defaults to root", func(t *testing.T) {
		secret, err := sm.GetIndividualSecret("", "DB_HOST", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		if secret != "localhost" {
			t.Errorf("expected: localhost, got: %s.", secret)
		}
	})

	t.Run("GetIndividualSecret not found", func(t *testing.T) {
		_, err := sm.GetIndividualSecret("/", "NONEXISTENT", "", map[string]string{})
		if err == nil {
			t.Fatalf("expected error but got nil")
		}
	})
}

func TestInfisicalGetIndividualSecretError(t *testing.T) {
	mock := newMockInfisicalClient(nil, fmt.Errorf("api error"))
	sm := NewInfisicalBackend(mock, "my-project", "dev")

	_, err := sm.GetIndividualSecret("/", "KEY", "", map[string]string{})
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
}

func TestInfisicalLogin(t *testing.T) {
	sm := NewInfisicalBackend(nil, "my-project", "dev")
	err := sm.Login()
	if err != nil {
		t.Fatalf("expected Login to return nil, got: %s", err)
	}
}
