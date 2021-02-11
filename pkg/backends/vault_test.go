package backends

import (
	"reflect"
	"testing"

	"github.com/IBM/argocd-vault-plugin/pkg/backends/auth"
	"github.com/IBM/argocd-vault-plugin/pkg/helpers"
)

func TestVaultLogin(t *testing.T) {
	cluster, role, secret := helpers.CreateTestAppRoleVault(t)
	defer cluster.Cleanup()

	vault := &Vault{
		VaultClient: cluster.Cores[0].Client,
	}

	t.Run("will authenticate with approle", func(t *testing.T) {
		vault.AuthType = &auth.AppRole{
			RoleID:   role,
			SecretID: secret,
		}

		err := vault.Login()
		if err != nil {
			t.Fatalf("expected no errors but got: %s", err)
		}
	})
}

func TestVaultGetSecrets(t *testing.T) {
	cluster, _, _ := helpers.CreateTestAppRoleVault(t)
	defer cluster.Cleanup()

	vault := &Vault{
		VaultClient: cluster.Cores[0].Client,
	}

	t.Run("will get data from vault with kv1", func(t *testing.T) {
		data, err := vault.GetSecrets("secret/foo", "1")
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
		data, err := vault.GetSecrets("kv/data/test", "2")
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

	t.Run("will throw an error if cant find secrets", func(t *testing.T) {
		_, err := vault.GetSecrets("kv/data/no_path", "2")
		if err == nil {
			t.Fatalf("expected an error but did not get an error")
		}

		expected := "Could not find secrets at path kv/data/no_path"

		if !reflect.DeepEqual(err.Error(), expected) {
			t.Errorf("expected: %s, got: %s.", expected, err.Error())
		}
	})

	t.Run("will throw an error if cant find secrets", func(t *testing.T) {
		_, err := vault.GetSecrets("kv/data/bad_test", "2")
		if err == nil {
			t.Fatalf("expected an error but did not get an error")
		}

		expected := "Could not get data from Vault, check that kv-v2 is the correct engine"

		if !reflect.DeepEqual(err.Error(), expected) {
			t.Errorf("expected: %s, got: %s.", expected, err.Error())
		}
	})

	t.Run("will throw an error if unsupported kv version", func(t *testing.T) {
		_, err := vault.GetSecrets("kv/data/test", "3")
		if err == nil {
			t.Fatalf("expected an error but did not get an error")
		}

		expected := "Unsupported kvVersion specified"

		if !reflect.DeepEqual(err.Error(), expected) {
			t.Errorf("expected: %s, got: %s.", expected, err.Error())
		}
	})

}
