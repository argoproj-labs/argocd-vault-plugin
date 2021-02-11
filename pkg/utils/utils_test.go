package utils_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/IBM/argocd-vault-plugin/pkg/helpers"
	"github.com/IBM/argocd-vault-plugin/pkg/utils"
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

	utils.SetToken(cluster.Cores[0].Client, "token")

	err := removeToken()
	if err != nil {
		t.Fatal(err)
	}
}
