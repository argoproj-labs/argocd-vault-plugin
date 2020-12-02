package vault

import (
	"reflect"
	"testing"
)

func TestAppRoleGetSecrets(t *testing.T) {
	ln, client := CreateTestVault(t)
	defer ln.Close()

	vc := &Client{
		VaultAPIClient: client,
	}

	appRole := AppRole{
		RoleID:   "role",
		SecretID: "secret",
		Client:   vc,
	}

	expected := map[string]interface{}{
		"secret": "bar",
	}

	data, err := appRole.GetSecrets("secret/foo")
	if err != nil {
		t.Fatalf("expected 0 errors but got: %s", err)
	}

	if !reflect.DeepEqual(data, expected) {
		t.Errorf("expected: %s, got: %s.", expected, data)
	}
}
