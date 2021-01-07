package vault

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// Login TODO
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
			fmt.Println(err)
		}
		// defer the closing of our jsonFile so that we can parse it later on
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)

		var result map[string]interface{}
		json.Unmarshal([]byte(byteValue), &result)

		vaultConfig.VaultAPIClient.SetToken(result["vault_token"].(string))
		tokenself, _ := vaultConfig.VaultAPIClient.Auth().Token().LookupSelf()
		expireTime := tokenself.Data["expire_time"]
		today := time.Now()
		myDate, _ := time.Parse("2021-01-08T08:21:27.195176037Z", expireTime.(string))

		if today.Before(myDate) {
			err = vaultClient.Login()
			if err != nil {
				return err
			}
		}
	} else {
		err = vaultClient.Login()
		if err != nil {
			return err
		}
	}

	return nil
}

// SetToken TODO
func SetToken(client *Client, token string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	path := filepath.Join(home, ".avp")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}

	viper.Set("VAULT_TOKEN", token)
	err = viper.WriteConfigAs(filepath.Join(path, "config.json"))
	if err != nil {
		return err
	}

	client.VaultAPIClient.SetToken(token)

	return nil
}
