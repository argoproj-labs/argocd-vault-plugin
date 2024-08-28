package vault_test

import (
	"bytes"
	"testing"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/auth/vault"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/helpers"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
)

func TestCertificateLogin(t *testing.T) {
	cluster, cert, key := helpers.CreateTestCertificateVault(t)
	defer cluster.Cleanup()

	certificateAuth := vault.NewCertificateAuth(cert, key, "")

	err := certificateAuth.Authenticate(cluster.Cores[0].Client)
	if err != nil {
		t.Fatalf("expected no errors but got: %s", err)
	}

	cachedToken, err := utils.ReadExistingToken()
	if err != nil {
		t.Fatalf("expected cached vault token but got: %s", err)
	}

	err = certificateAuth.Authenticate(cluster.Cores[0].Client)
	if err != nil {
		t.Fatalf("expected no errors but got: %s", err)
	}

	newCachedToken, err := utils.ReadExistingToken()
	if err != nil {
		t.Fatalf("expected cached vault token but got: %s", err)
	}

	if bytes.Compare(cachedToken, newCachedToken) != 0 {
		t.Fatalf("expected same token %s but got %s", cachedToken, newCachedToken)
	}
}
