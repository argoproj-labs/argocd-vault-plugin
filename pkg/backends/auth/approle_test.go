package auth_test

import (
	"testing"

	"github.com/IBM/argocd-vault-plugin/pkg/backends/auth"
	"github.com/IBM/argocd-vault-plugin/pkg/helpers"
)

func TestAppRoleLogin(t *testing.T) {
	cluster, role, secret := helpers.CreateTestAppRoleVault(t)
	defer cluster.Cleanup()

	appRole := auth.AppRole{
		RoleID:   role,
		SecretID: secret,
	}

	err := appRole.Authenticate(cluster.Cores[0].Client)
	if err != nil {
		t.Fatalf("expected no errors but got: %s", err)
	}
}
