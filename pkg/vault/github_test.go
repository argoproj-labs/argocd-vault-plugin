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

	t.Run("will get data from vault with kv1", func(t *testing.T) {
		data, err := github.GetSecrets("secret/foo", "1")
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"secret": "bar",
		}

		if !reflect.DeepEqual(data, expected) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}
	})

	t.Run("will get data from vault with kv2", func(t *testing.T) {
		data, err := github.GetSecrets("kv/data/test", "2")
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"hello": "world",
		}

		if !reflect.DeepEqual(data, expected) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}
	})
}
