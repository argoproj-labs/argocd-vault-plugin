package vault_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/auth/vault"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/helpers"
)

const saPath = "/tmp/avp/kubernetes.io/serviceaccount"

func writeK8sToken() error {
	err := os.MkdirAll(saPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Could not create directory: %s", err.Error())
	}

	data := []byte("123456")
	err = ioutil.WriteFile(filepath.Join(saPath, "token"), data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func removeK8sToken() error {
	err := os.RemoveAll("/tmp/avp")
	if err != nil {
		return err
	}
	return nil
}

// Need to find a way to mock GitHub Auth within Vault
func TestKubernetesAuth(t *testing.T) {
	cluster := helpers.CreateTestAuthVault(t)
	defer cluster.Cleanup()

	err := writeK8sToken()
	if err != nil {
		t.Fatalf("error writing token: %s", err)
	}

	k8s := vault.NewK8sAuth("role", "", string(filepath.Join(saPath, "token")))

	err = k8s.Authenticate(cluster.Cores[0].Client)
	if err != nil {
		t.Fatalf("expected no errors but got: %s", err)
	}

	err = removeK8sToken()
	if err != nil {
		fmt.Println(err)
	}
}
