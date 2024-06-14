package vault_test

import (
	"bytes"
	"testing"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/auth/vault"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/helpers"
)

func TestUserPassLogin(t *testing.T) {
	cluster, username, password := helpers.CreateTestUserPassVault(t)
	defer cluster.Cleanup()

	userpass := vault.NewUserPassAuth(username, password, "")

	if err := userpass.Authenticate(cluster.Cores[0].Client); err != nil {
		t.Fatalf("expected no errors but got: %s", err)
	}

	cachedToken, err := utils.ReadExistingToken(cluster.Cores[0].Client)
	if err != nil {
		t.Fatalf("expected cached vault token but got: %s", err)
	}

	err = userpass.Authenticate(cluster.Cores[0].Client)
	if err != nil {
		t.Fatalf("expected no errors but got: %s", err)
	}

	newCachedToken, err := utils.ReadExistingToken(cluster.Cores[0].Client)
	if err != nil {
		t.Fatalf("expected cached vault token but got: %s", err)
	}

	if bytes.Compare(cachedToken, newCachedToken) != 0 {
		t.Fatalf("expected same token %s but got %s", cachedToken, newCachedToken)
	}
}
