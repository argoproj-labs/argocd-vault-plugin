package vault_test

import (
	"bytes"
	"testing"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/auth/vault"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/helpers"
)

// Need to find a way to mock GitHub Auth within Vault
func TestGithubLogin(t *testing.T) {
	cluster := helpers.CreateTestAuthVault(t)
	defer cluster.Cleanup()

	github := vault.NewGithubAuth("123", "", "default")

	err := github.Authenticate(cluster.Cores[0].Client)
	if err != nil {
		t.Fatalf("expected no errors but got: %s", err)
	}

	cachedToken, err := utils.ReadExistingToken("github_default")
	if err != nil {
		t.Fatalf("expected cached vault token but got: %s", err)
	}

	err = github.Authenticate(cluster.Cores[0].Client)
	if err != nil {
		t.Fatalf("expected no errors but got: %s", err)
	}

	newCachedToken, err := utils.ReadExistingToken("github_default")
	if err != nil {
		t.Fatalf("expected cached vault token but got: %s", err)
	}

	if bytes.Compare(cachedToken, newCachedToken) != 0 {
		t.Fatalf("expected same token %s but got %s", cachedToken, newCachedToken)
	}

	// We create a new connection to a specific namespace and create a different cache
	namespaceCluster := helpers.CreateTestAuthVault(t)
	defer namespaceCluster.Cleanup()

	namespaceGithub := vault.NewGithubAuth("123", "", "my-other-namespace")

	err = namespaceGithub.Authenticate(namespaceCluster.Cores[0].Client)
	if err != nil {
		t.Fatalf("expected no errors but got: %s", err)
	}

	_, err = utils.ReadExistingToken("github_my-other-namespace")
	if err != nil {
		t.Fatalf("expected cached vault token but got: %s", err)
	}
}
