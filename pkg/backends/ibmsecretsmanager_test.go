package backends_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/IBM/argocd-vault-plugin/pkg/backends"
	"github.com/IBM/argocd-vault-plugin/pkg/helpers"
)

func TestSecretsManagerGetSecrets(t *testing.T) {
	ln, client, _ := helpers.CreateTestVault(t)
	defer ln.Close()

	sm := backends.IBMSecretsManager{
		IBMCloudAPIKey: "token",
		VaultClient:    client,
	}

	expected := map[string]interface{}{
		"secret":  "value",
		"secret2": "value2",
	}

	data, err := sm.GetSecrets("secret/ibm/arbitrary/groups/1", map[string]string{})
	if err != nil {
		t.Fatalf("expected 0 errors but got: %s", err)
	}

	if !reflect.DeepEqual(data, expected) {
		t.Errorf("expected: %s, got: %s.", expected, data)
	}
}

func TestSecretsmanagerGetSecretsFail(t *testing.T) {
	ln, client, _ := helpers.CreateTestVault(t)
	defer ln.Close()

	sm := backends.IBMSecretsManager{
		IBMCloudAPIKey: "token",
		VaultClient:    client,
	}

	_, err := sm.GetSecrets("secret/ibm/arbitrary/groups/3", map[string]string{})

	expected := fmt.Sprintf("Could not find secrets at path %s", "secret/ibm/arbitrary/groups/3")

	if !reflect.DeepEqual(err.Error(), expected) {
		t.Errorf("expected: %s, got: %s.", expected, err.Error())
	}
}
