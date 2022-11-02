package utils_test

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/helpers"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
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

func TestCheckExistingToken(t *testing.T) {
	ln, client, roottoken := helpers.CreateTestVault(t)
	defer ln.Close()

	t.Run("will set token if valid", func(t *testing.T) {
		err := writeToken(roottoken)
		if err != nil {
			t.Fatal(err)
		}

		err = utils.CheckExistingToken(client)
		if err != nil {
			t.Fatal(err)
		}

		token := client.Token()
		if !reflect.DeepEqual(token, roottoken) {
			t.Errorf("expected: %s, got: %s.", roottoken, token)
		}

		err = removeToken()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("will throw an error if no toekn", func(t *testing.T) {
		ln, client, _ := helpers.CreateTestVault(t)
		defer ln.Close()

		err := utils.CheckExistingToken(client)
		if err == nil {
			t.Fatal(err)
		}

		dir, _ := os.UserHomeDir()
		expected := fmt.Sprintf("stat %s/.avp/config.json: no such file or directory", dir)
		if err.Error() != expected {
			t.Errorf("expected: %s, got: %s.", expected, err.Error())
		}
	})

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

func TestDefaultHTTPClient(t *testing.T) {
	expectedClient := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	client := utils.DefaultHttpClient()

	if !reflect.DeepEqual(client, expectedClient) {
		t.Errorf("expected: %v, got: %v.", expectedClient, client)
	}
}
