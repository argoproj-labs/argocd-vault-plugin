package vault

import (
	"reflect"
	"testing"
)

func TestSecretManagerGetSecrets(t *testing.T) {
	ln, client := CreateTestVault(t)
	defer ln.Close()

	vc := &Client{
		PathPrefix:     "secret",
		VaultAPIClient: client,
	}

	sm := SecretManager{
		IAMToken: "token",
		Client:   vc,
	}

	expected := map[string]interface{}{
		"secret": "bar",
	}

	data, err := sm.GetSecrets("/foo")
	if err != nil {
		t.Fatalf("expected 0 errors but got: %s", err)
	}

	if !reflect.DeepEqual(data, expected) {
		t.Errorf("expected: %s, got: %s.", expected, data)
	}
}
