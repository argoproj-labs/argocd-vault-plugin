package backends_test

import (
	"reflect"
	"testing"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/auth/vault"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/backends"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/helpers"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/types"
)

func TestVaultLogin(t *testing.T) {
	cluster, roleID, secretID := helpers.CreateTestAppRoleVault(t)
	defer cluster.Cleanup()

	backend := &backends.Vault{
		VaultClient: cluster.Cores[0].Client,
	}

	t.Run("will authenticate with approle", func(t *testing.T) {
		backend.AuthType = vault.NewAppRoleAuth(roleID, secretID, "")

		err := backend.Login()
		if err != nil {
			t.Fatalf("expected no errors but got: %s", err)
		}
	})
}

func TestVaultGetSecrets(t *testing.T) {
	cluster, roleID, secretID := helpers.CreateTestAppRoleVault(t)
	defer cluster.Cleanup()

	auth := vault.NewAppRoleAuth(roleID, secretID, "")
	backend := backends.NewVaultBackend(auth, cluster.Cores[0].Client, "")

	t.Run("will get data from vault with kv1", func(t *testing.T) {
		annotations := map[string]string{
			types.VaultKVVersionAnnotation: "1",
		}
		data, err := backend.GetSecrets("secret/foo", "", annotations)
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
		annotations := map[string]string{
			types.VaultKVVersionAnnotation: "2",
		}
		data, err := backend.GetSecrets("kv/data/test", "", annotations)
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

	t.Run("Vault GetIndividualSecret", func(t *testing.T) {
		secret, err := backend.GetIndividualSecret("kv/data/test", "hello", "", map[string]string{
			types.VaultKVVersionAnnotation: "2",
		})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := "world"

		if !reflect.DeepEqual(expected, secret) {
			t.Errorf("expected: %s, got: %s", expected, secret)
		}
	})

	t.Run("will honor version with kv2", func(t *testing.T) {
		annotations := map[string]string{
			types.VaultKVVersionAnnotation: "2",
		}

		// Get version 1 when there are 2 versions
		data, err := backend.GetSecrets("kv/data/versioned", "1", annotations)
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"secret": "version1",
		}

		if !reflect.DeepEqual(data, expected) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}

		// Get version 2 (latest)
		data, err = backend.GetSecrets("kv/data/versioned", "", annotations)
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected = map[string]interface{}{
			"secret": "version2",
		}

		if !reflect.DeepEqual(data, expected) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}
	})

	t.Run("will throw an error if cant find secrets", func(t *testing.T) {
		annotations := map[string]string{
			types.VaultKVVersionAnnotation: "2",
		}
		_, err := backend.GetSecrets("kv/data/no_path", "", annotations)
		if err == nil {
			t.Fatalf("expected an error but did not get an error")
		}

		expected := "Could not find secrets at path kv/data/no_path"

		if !reflect.DeepEqual(err.Error(), expected) {
			t.Errorf("expected: %s, got: %s.", expected, err.Error())
		}
	})

	t.Run("will throw an error if cant find secrets", func(t *testing.T) {
		annotations := map[string]string{
			types.VaultKVVersionAnnotation: "2",
		}
		_, err := backend.GetSecrets("secret/bad_test", "", annotations)
		if err == nil {
			t.Fatalf("expected an error but did not get an error")
		}

		expected := "Could not get data from Vault, check that kv-v2 is the correct engine"

		if !reflect.DeepEqual(err.Error(), expected) {
			t.Errorf("expected: %s, got: %s.", expected, err.Error())
		}
	})

	t.Run("will throw an error if unsupported kv version", func(t *testing.T) {
		annotations := map[string]string{
			types.VaultKVVersionAnnotation: "3",
		}
		_, err := backend.GetSecrets("kv/data/test", "", annotations)
		if err == nil {
			t.Fatalf("expected an error but did not get an error")
		}

		expected := "Unsupported kvVersion specified"

		if !reflect.DeepEqual(err.Error(), expected) {
			t.Errorf("expected: %s, got: %s.", expected, err.Error())
		}
	})

}
