package utils

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
)

// CheckExistingToken takes a VaultType interface and logs in, while writting the config file
// And setting the token in the client
func CheckExistingToken(vaultClient *api.Client) error {
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

	vaultClient.SetToken(result["vault_token"].(string))
	_, err = vaultClient.Auth().Token().LookupSelf()
	if err != nil {
		return err
	}

	return nil
}

// SetToken attmepts to set the vault token on the vault api client
// and then attempts to write that token to a file to be used later
// If this method fails we do not want to stop the process
func SetToken(vaultClient *api.Client, token string) error {
	// We want to set the token first
	vaultClient.SetToken(token)

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

func DefaultHttpClient() *http.Client {
	var tlsClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	var transport http.RoundTripper = &http.Transport{
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     tlsClientConfig,
	}

	var httpClient = &http.Client{
		Timeout:   60 * time.Second,
		Transport: transport,
	}

	return httpClient
}

func VerboseToStdErr(format string, message ...interface{}) {
	if viper.GetBool("verboseOutput") {
		log.Printf(fmt.Sprintf("%s\n", format), message...)
	}
}
