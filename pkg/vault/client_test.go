package vault_test

import (
	"reflect"
	"testing"

	"github.com/IBM/argocd-vault-plugin/pkg/helpers"
	"github.com/IBM/argocd-vault-plugin/pkg/vault"
	"github.com/hashicorp/vault/api"
)

func TestVaultRead(t *testing.T) {
	ln, client := helpers.CreateTestVault(t)
	defer ln.Close()

	vc := &vault.Client{
		VaultAPIClient: client,
	}

	t.Run("will get data from vault", func(t *testing.T) {
		secret, err := vc.Read("secret/foo")
		if err != nil {
			t.Error(err)
		}

		if secret.Data["secret"] != "bar" {
			t.Errorf("expected: %s, got: %s.", "bar", secret.Data["secret"])
		}
	})

	t.Run("will get empty map if no path exists", func(t *testing.T) {
		secret, err := vc.Read("secret/bar")
		if err != nil {
			t.Error(err)
		}

		expected := &api.Secret{}
		if !reflect.DeepEqual(secret, expected) {
			t.Errorf("expected: %v, got: %v.", expected, secret)
		}
	})

	t.Run("will write to path", func(t *testing.T) {
		payload := map[string]interface{}{
			"new_secret": "value",
		}
		_, err := vc.Write("secret/bar", payload)
		if err != nil {
			t.Error(err)
		}

		secret, err := vc.Read("secret/bar")
		if err != nil {
			t.Error(err)
		}

		data := secret.Data

		if !reflect.DeepEqual(data, payload) {
			t.Errorf("expected: %s, got: %s.", payload, data)
		}
	})
}
