package vault_test

import (
	"bytes"
	"fmt"
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

	cachedToken, err := utils.ReadExistingToken(cluster.Cores[0].Client, fmt.Sprintf("userpass_%s", username))

	if err != nil {
		t.Fatalf("expected cached vault token but got: %s", err)
	}

	err = userpass.Authenticate(cluster.Cores[0].Client)
	if err != nil {
		t.Fatalf("expected no errors but got: %s", err)
	}

	newCachedToken, err := utils.ReadExistingToken(cluster.Cores[0].Client, fmt.Sprintf("userpass_%s", username))

	if err != nil {
		t.Fatalf("expected cached vault token but got: %s", err)
	}

	if bytes.Compare(cachedToken, newCachedToken) != 0 {
		t.Fatalf("expected same token %s but got %s", cachedToken, newCachedToken)
	}

	// We create a new connection with a different approle and create a different cache
	secondCluster, secondUsername, secondPassword := helpers.CreateTestUserPassVault(t)
	defer secondCluster.Cleanup()

	secondUserpass := vault.NewUserPassAuth(secondUsername, secondPassword, "")

	err = secondUserpass.Authenticate(secondCluster.Cores[0].Client)
	if err != nil {
		t.Fatalf("expected no errors but got: %s", err)
	}

	secondCachedToken, err := utils.ReadExistingToken(cluster.Cores[0].Client, fmt.Sprintf("userpass_%s", secondUsername))
	if err != nil {
		t.Fatalf("expected cached vault token but got: %s", err)
	}

	// Both cache should be different
	if bytes.Compare(cachedToken, secondCachedToken) == 0 {
		t.Fatalf("expected different tokens but got %s", secondCachedToken)
	}
}
