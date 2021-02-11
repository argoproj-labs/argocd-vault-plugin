package backends_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/IBM/argocd-vault-plugin/pkg/backends"
	"github.com/IBM/argocd-vault-plugin/pkg/helpers"
)

func TestSecretManagerGetSecrets(t *testing.T) {
	ln, client := helpers.CreateTestVault(t)
	defer ln.Close()

	sm := backends.SecretManager{
		IBMCloudAPIKey: "token",
		VaultClient:    client,
	}

	expected := map[string]interface{}{
		"secret":  "value",
		"secret2": "value2",
	}

	data, err := sm.GetSecrets("secret/ibm/arbitrary/groups/1", "")
	if err != nil {
		t.Fatalf("expected 0 errors but got: %s", err)
	}

	if !reflect.DeepEqual(data, expected) {
		t.Errorf("expected: %s, got: %s.", expected, data)
	}
}

func TestSecretManagerGetSecretsFail(t *testing.T) {
	ln, client := helpers.CreateTestVault(t)
	defer ln.Close()

	sm := backends.SecretManager{
		IBMCloudAPIKey: "token",
		VaultClient:    client,
	}

	_, err := sm.GetSecrets("secret/ibm/arbitrary/groups/3", "")

	expected := fmt.Sprintf("Could not find secrets at path %s", "secret/ibm/arbitrary/groups/3")

	if !reflect.DeepEqual(err.Error(), expected) {
		t.Errorf("expected: %s, got: %s.", expected, err.Error())
	}
}
