package utils

import (
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
)

func PurgeTokenCache() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	avpConfigFolderPath := filepath.Join(home, ".avp")

	err = os.RemoveAll(avpConfigFolderPath)
	if err != nil {
		return err
	}
	return nil
}

func GetConfigFileName(vaultClient *api.Client, identifier string) string {
	var config_prefix = identifier
	var config_ext = ".json"
	var config_name = "_" + vaultClient.Namespace()

	hasher := sha1.New()
	hasher.Write([]byte(vaultClient.Address()))
	var config_addr_name = "_" + base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	config := config_prefix + config_addr_name + config_name + config_ext

	return config
}

func ReadExistingToken(vaultClient *api.Client, identifier string) ([]byte, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	avpConfigPath := filepath.Join(home, ".avp", GetConfigFileName(vaultClient, identifier))
	if _, err := os.Stat(avpConfigPath); err != nil {
		return nil, err
	}

	// Open our jsonFile
	jsonFile, err := os.Open(avpConfigPath)
	if err != nil {
		return nil, err
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	return byteValue, nil
}

// LoginWithCachedToken takes a VaultType interface, the auth types and an identifier of the token
// It then tries to log in with the matching previously cached token,
// And sets the token in the client
func LoginWithCachedToken(vaultClient *api.Client, identifier string) error {
	if viper.GetBool("disableCache") {
		return fmt.Errorf("Token cache feature is disabled")
	} else {
		byteValue, err := ReadExistingToken(vaultClient, identifier)
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
}

// SetToken attmepts to set the vault token on the vault api client
// and then attempts to write that token to a file to be used later
// If this method fails we do not want to stop the process
func SetToken(vaultClient *api.Client, identifier string, token string) error {
	// We want to set the token first
	vaultClient.SetToken(token)

	if viper.GetBool("disableCache") {
		return fmt.Errorf("Token cache feature is disabled")
	} else {
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
			"vault_addr":      os.Getenv("VAULT_ADDR"),
			"vault_namespace": os.Getenv("VAULT_NAMESPACE"),
			"vault_token":     token,
		}
		file, err := json.MarshalIndent(data, "", " ")
		if err != nil {
			return fmt.Errorf("Could not marshal token data: %s", err.Error())
		}
		err = os.WriteFile(filepath.Join(path, GetConfigFileName(vaultClient, identifier)), file, 0644)
		if err != nil {
			return fmt.Errorf("Could not write token to file, will need to login to Vault on subsequent runs: %s", err.Error())
		}

		return nil
	}
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
