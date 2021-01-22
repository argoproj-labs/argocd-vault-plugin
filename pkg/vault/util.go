package vault

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Login takes a VaultType interface and logs in, while writting the config file
// And setting the token in the client
func Login(vaultClient VaultType, vaultConfig *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	avpConfigPath := filepath.Join(home, ".avp", "config.json")
	if _, err := os.Stat(avpConfigPath); err == nil {
		// Open our jsonFile
		jsonFile, err := os.Open(avpConfigPath)
		if err != nil {
			return err
		}
		// defer the closing of our jsonFile so that we can parse it later on
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)

		var result map[string]interface{}
		json.Unmarshal([]byte(byteValue), &result)

		vaultConfig.VaultAPIClient.SetToken(result["vault_token"].(string))
		_, err = vaultConfig.VaultAPIClient.Auth().Token().LookupSelf()
		if err != nil {
			err = vaultClient.Login()
			if err != nil {
				return err
			}
		}

		return nil
	}

	err = vaultClient.Login()
	if err != nil {
		return err
	}

	return nil
}

// SetToken TODO
func SetToken(client *Client, token string) {
	// We want to set the token first
	client.VaultAPIClient.SetToken(token)

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Could not access home directory, will need to login to Vault on subsequent runs: %s", err.Error())
	}

	path := filepath.Join(home, ".avp")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}

	data := map[string]interface{}{
		"vault_token": token,
	}
	file, _ := json.MarshalIndent(data, "", " ")
	err = ioutil.WriteFile(filepath.Join(path, "config.json"), file, 0644)
	if err != nil {
		fmt.Printf("Could not write token to file, will need to login to Vault on subsequent runs: %s", err.Error())
	}
}
