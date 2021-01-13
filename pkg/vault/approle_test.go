package vault_test

import (
	"reflect"
	"testing"

	"github.com/IBM/argocd-vault-plugin/pkg/helpers"
	"github.com/IBM/argocd-vault-plugin/pkg/vault"
)

func TestAppRoleLogin(t *testing.T) {
	cluster, role, secret := helpers.CreateTestAppRoleVault(t)
	defer cluster.Cleanup()

	vc := &vault.Client{
		VaultAPIClient: cluster.Cores[0].Client,
	}

	appRole := vault.AppRole{
		RoleID:   role,
		SecretID: secret,
		Client:   vc,
	}

	err := appRole.Login()
	if err != nil {
		t.Fatalf("expected no errors but got: %s", err)
	}
}

func TestAppRoleGetSecrets(t *testing.T) {
	cluster, role, secret := helpers.CreateTestAppRoleVault(t)
	defer cluster.Cleanup()

	vc := &vault.Client{
		VaultAPIClient: cluster.Cores[0].Client,
	}

	appRole := vault.AppRole{
		RoleID:   role,
		SecretID: secret,
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
