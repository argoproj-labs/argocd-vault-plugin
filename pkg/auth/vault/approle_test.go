package vault_test

import (
	"testing"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/auth/vault"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/helpers"
)

func TestAppRoleLogin(t *testing.T) {
	cluster, roleID, secretID := helpers.CreateTestAppRoleVault(t)
	defer cluster.Cleanup()

	appRole := vault.NewAppRoleAuth(roleID, secretID, "")

	err := appRole.Authenticate(cluster.Cores[0].Client)
	if err != nil {
		t.Fatalf("expected no errors but got: %s", err)
	}
}
