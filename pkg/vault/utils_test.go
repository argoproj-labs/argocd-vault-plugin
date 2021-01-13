package vault_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/IBM/argocd-vault-plugin/pkg/helpers"
	"github.com/IBM/argocd-vault-plugin/pkg/vault"
)

func writeToken(token string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".avp")
	os.Mkdir(path, 0755)
	data := map[string]interface{}{
		"vault_token": token,
	}
	file, _ := json.MarshalIndent(data, "", " ")
	err = ioutil.WriteFile(filepath.Join(path, "config.json"), file, 0644)
	if err != nil {
		return err
	}
	return nil
}

func removeToken() error {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".avp")
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	return nil
}

func readToken() interface{} {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".avp", "config.json")
	dat, _ := ioutil.ReadFile(path)
	var result map[string]interface{}
	json.Unmarshal([]byte(dat), &result)
	return result["vault_token"]
}

func TestSetToken(t *testing.T) {
	cluster, _, _ := helpers.CreateTestAppRoleVault(t)
	defer cluster.Cleanup()

	vc := &vault.Client{
		VaultAPIClient: cluster.Cores[0].Client,
	}

	err := vault.SetToken(vc, "token")
	if err != nil {
		t.Errorf("expected token to be written, got: %s.", err)
	}

	err = removeToken()
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoginWithNoToken(t *testing.T) {
	cluster, role, secret := helpers.CreateTestAppRoleVault(t)
	defer cluster.Cleanup()

	vc := &vault.Client{
		VaultAPIClient: cluster.Cores[0].Client,
	}

	cf := &vault.Config{
		Address:    "address",
		PathPrefix: "prefix",
		Type: &vault.AppRole{
			RoleID:   role,
			SecretID: secret,
			Client:   vc,
		},
		Client: vc,
	}

	err := removeToken()
	if err != nil {
		t.Fatal(err)
	}

	err = vault.Login(cf.Type, cf)
	if err != nil {
		t.Errorf("expected: %s, got: %s.", "", err)
	}

	token := readToken()
	if token == "" {
		t.Errorf("expected a vault token, got: %s.", token.(string))
	}

	err = removeToken()
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoginWithOldToken(t *testing.T) {
	cluster, role, secret := helpers.CreateTestAppRoleVault(t)
	defer cluster.Cleanup()

	vc := &vault.Client{
		VaultAPIClient: cluster.Cores[0].Client,
	}

	cf := &vault.Config{
		Address:    "address",
		PathPrefix: "prefix",
		Type: &vault.AppRole{
			RoleID:   role,
			SecretID: secret,
			Client:   vc,
		},
		Client: vc,
	}

	err := writeToken("token")
	if err != nil {
		t.Fatal(err)
	}

	err = vault.Login(cf.Type, cf)
	if err != nil {
		t.Errorf("expected: %s, got: %s.", "", err)
	}

	token := readToken()
	if token == "" {
		t.Errorf("expected a vault token, got: %s.", token.(string))
	}

	err = removeToken()
	if err != nil {
		t.Fatal(err)
	}
}
