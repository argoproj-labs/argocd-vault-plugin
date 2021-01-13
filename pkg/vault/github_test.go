package vault_test

import (
	"reflect"
	"testing"

	"github.com/IBM/argocd-vault-plugin/pkg/helpers"
	"github.com/IBM/argocd-vault-plugin/pkg/vault"
)

func TestGithubGetSecrets(t *testing.T) {
	ln, client := helpers.CreateTestVault(t)
	defer ln.Close()

	vc := &vault.Client{
		VaultAPIClient: client,
	}

	github := vault.Github{
		AccessToken: "token",
		Client:      vc,
	}

	expected := map[string]interface{}{
		"secret": "bar",
	}

	data, err := github.GetSecrets("secret/foo")
	if err != nil {
		t.Fatalf("expected 0 errors but got: %s", err)
	}

	if !reflect.DeepEqual(data, expected) {
		t.Errorf("expected: %s, got: %s.", expected, data)
	}
}
