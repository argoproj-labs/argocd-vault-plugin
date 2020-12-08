package vault

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSecretManagerGetSecrets(t *testing.T) {
	ln, client := CreateTestVault(t)
	defer ln.Close()

	vc := &Client{
		VaultAPIClient: client,
	}

	sm := SecretManager{
		IBMCloudAPIKey: "token",
		Client:         vc,
	}

	expected := map[string]interface{}{
		"secret":  "value",
		"secret2": "value2",
	}

	data, err := sm.GetSecrets("secret/ibm/arbitrary/groups/1")
	if err != nil {
		t.Fatalf("expected 0 errors but got: %s", err)
	}

	if !reflect.DeepEqual(data, expected) {
		t.Errorf("expected: %s, got: %s.", expected, data)
	}
}

func TestSecretManagerGetSecretsFail(t *testing.T) {
	ln, client := CreateTestVault(t)
	defer ln.Close()

	vc := &Client{
		VaultAPIClient: client,
	}

	sm := SecretManager{
		IBMCloudAPIKey: "token",
		Client:         vc,
	}

	_, err := sm.GetSecrets("secret/ibm/arbitrary/groups/3")
	if err == nil {
		t.Fatalf("expected an error but did not recieve one")
	}

	expected := fmt.Sprintf("Could not find secrets at path %s", "secret/ibm/arbitrary/groups/3")

	if !reflect.DeepEqual(err.Error(), expected) {
		t.Errorf("expected: %s, got: %s.", expected, err.Error())
	}
}
