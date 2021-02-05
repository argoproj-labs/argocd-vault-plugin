package vault

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// CheckExistingToken takes a VaultType interface and logs in, while writting the config file
// And setting the token in the client
func CheckExistingToken(vaultClient VaultType, vaultConfig *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	avpConfigPath := filepath.Join(home, ".avp", "config.json")
	if _, err := os.Stat(avpConfigPath); err != nil {
		return err
	}

	// Open our jsonFile
	jsonFile, err := os.Open(avpConfigPath)
	if err != nil {
		return err
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	err = json.Unmarshal([]byte(byteValue), &result)
	if err != nil {
		return err
	}

	vaultConfig.VaultAPIClient.SetToken(result["vault_token"].(string))
	_, err = vaultConfig.VaultAPIClient.Auth().Token().LookupSelf()
	if err != nil {
		return err
	}

	return nil
}

// SetToken TODO
func SetToken(client *Client, token string) error {
	// We want to set the token first
	client.VaultAPIClient.SetToken(token)

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("Could not access home directory: %s", err.Error())
	}

	path := filepath.Join(home, ".avp")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, 0755)
		if err != nil {
			return fmt.Errorf("Could not create avp directory: %s", err.Error())
		}
	}

	data := map[string]interface{}{
		"vault_token": token,
	}
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return fmt.Errorf("Could not marshal token data: %s", err.Error())
	}
	err = ioutil.WriteFile(filepath.Join(path, "config.json"), file, 0644)
	if err != nil {
		return fmt.Errorf("Could not write token to file, will need to login to Vault on subsequent runs: %s", err.Error())
	}

	return nil
}

// ReadVaultSecret calls the vault client and returns a data based on the KV Version
func ReadVaultSecret(client Client, path, kvVersion string) (map[string]interface{}, error) {
	secret, err := client.Read(path)
	if err != nil {
		return nil, err
	}

	if secret == nil {
		return map[string]interface{}{}, nil
	}

	if kvVersion == "2" {
		if _, ok := secret.Data["data"]; ok {
			return secret.Data["data"].(map[string]interface{}), nil
		}
		return nil, errors.New("Could not get data from Vault, check that kv-v2 is the correct engine")
	}

	if kvVersion == "1" {
		return secret.Data, nil
	}

	return nil, errors.New("Unsupported kvVersion specified")
}
