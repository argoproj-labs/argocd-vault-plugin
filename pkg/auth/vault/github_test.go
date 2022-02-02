package vault_test

import (
	"testing"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/auth/vault"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/helpers"
)

// Need to find a way to mock GitHub Auth within Vault
func TestGithubLogin(t *testing.T) {
	cluster := helpers.CreateTestAuthVault(t)
	defer cluster.Cleanup()

	github := vault.NewGithubAuth("123", "")

	err := github.Authenticate(cluster.Cores[0].Client)
	if err != nil {
		t.Fatalf("expected no errors but got: %s", err)
	}
}
