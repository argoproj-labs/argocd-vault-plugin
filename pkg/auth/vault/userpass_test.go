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

	userpass := vault.NewUserPassAuth(username, password, "", "default")

	if err := userpass.Authenticate(cluster.Cores[0].Client); err != nil {
		t.Fatalf("expected no errors but got: %s", err)
	}

	cachedToken, err := utils.ReadExistingToken(fmt.Sprintf("userpass_default_%s", username))
	if err != nil {
		t.Fatalf("expected cached vault token but got: %s", err)
	}

	err = userpass.Authenticate(cluster.Cores[0].Client)
	if err != nil {
		t.Fatalf("expected no errors but got: %s", err)
	}

	newCachedToken, err := utils.ReadExistingToken(fmt.Sprintf("userpass_default_%s", username))
	if err != nil {
		t.Fatalf("expected cached vault token but got: %s", err)
	}

	if bytes.Compare(cachedToken, newCachedToken) != 0 {
		t.Fatalf("expected same token %s but got %s", cachedToken, newCachedToken)
	}

	// We create a new connection with a different approle and create a different cache
	secondCluster, secondUsername, secondPassword := helpers.CreateTestUserPassVault(t)
	defer secondCluster.Cleanup()

	secondUserpass := vault.NewUserPassAuth(secondUsername, secondPassword, "", "default")

	err = secondUserpass.Authenticate(secondCluster.Cores[0].Client)
	if err != nil {
		t.Fatalf("expected no errors but got: %s", err)
	}

	secondCachedToken, err := utils.ReadExistingToken(fmt.Sprintf("userpass_default_%s", secondUsername))
	if err != nil {
		t.Fatalf("expected cached vault token but got: %s", err)
	}

	// Both cache should be different
	if bytes.Compare(cachedToken, secondCachedToken) == 0 {
		t.Fatalf("expected different tokens but got %s", secondCachedToken)
	}

	// We create a new connection to a specific namespace and create a different cache
	namespaceCluster, namespaceUsername, namespacePassword := helpers.CreateTestUserPassVault(t)
	defer namespaceCluster.Cleanup()

	namespaceUserpass := vault.NewUserPassAuth(namespaceUsername, namespacePassword, "", "my-other-namespace")

	err = namespaceUserpass.Authenticate(namespaceCluster.Cores[0].Client)
	if err != nil {
		t.Fatalf("expected no errors but got: %s", err)
	}

	_, err = utils.ReadExistingToken(fmt.Sprintf("userpass_my-other-namespace_%s", namespaceUsername))
	if err != nil {
		t.Fatalf("expected cached vault token but got: %s", err)
	}
}
