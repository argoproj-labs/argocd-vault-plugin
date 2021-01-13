package vault_test

import (
	"reflect"
	"testing"

	"github.com/IBM/argocd-vault-plugin/pkg/helpers"
	"github.com/IBM/argocd-vault-plugin/pkg/vault"
)

func TestVaultRead(t *testing.T) {
	ln, client := helpers.CreateTestVault(t)
	defer ln.Close()

	vc := &vault.Client{
		VaultAPIClient: client,
	}

	t.Run("will get data from vault", func(t *testing.T) {
		data, err := vc.Read("secret/foo")
		if err != nil {
			t.Error(err)
		}

		if data["secret"] != "bar" {
			t.Errorf("expected: %s, got: %s.", "bar", data["secret"])
		}
	})

	t.Run("will get empty map if no path exists", func(t *testing.T) {
		data, err := vc.Read("secret/bar")
		if err != nil {
			t.Error(err)
		}

		expected := map[string]interface{}{}
		if !reflect.DeepEqual(data, expected) {
			t.Errorf("expected: %s, got: %s.", expected, data)
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

		data, err := vc.Read("secret/bar")
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(data, payload) {
			t.Errorf("expected: %s, got: %s.", payload, data)
		}
	})
}
