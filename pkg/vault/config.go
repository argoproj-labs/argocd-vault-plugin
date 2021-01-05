package vault

import (
	"errors"
	"net/http"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
)

// Config TODO
type Config struct {
	Address    string
	PathPrefix string
	Type       VaultType
}

func setupViper() {
	viper.BindEnv("VaultType", "AVP_TYPE")
	viper.BindEnv("Address", "AVP_VAULT_ADDR")
	viper.BindEnv("PathPrefix", "AVP_PATH_PREFIX")
	viper.BindEnv("AuthType", "AVP_AUTH_TYPE")
	viper.BindEnv("RoleID", "AVP_ROLE_ID")
	viper.BindEnv("SecretID", "AVP_SECRET_ID")
	viper.BindEnv("GithubToken", "AVP_GITHUB_TOKEN")
	viper.BindEnv("IBMCloudAPIKey", "AVP_IBM_API_KEY")
}

// NewConfig returns a new Config struct
func NewConfig() (*Config, error) {
	setupViper()

	config := &Config{
		Address:    viper.GetString("Address"),
		PathPrefix: viper.GetString("PathPrefix"),
	}

	var httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	apiClient, err := api.NewClient(&api.Config{Address: viper.GetString("Address"), HttpClient: httpClient})
	if err != nil {
		return nil, err
	}

	client := &Client{
		VaultAPIClient: apiClient,
	}

	auth := viper.GetString("AuthType")

	switch viper.GetString("VaultType") {
	case "vault":
		switch auth {
		case "approle":
			if viper.IsSet("RoleID") && viper.IsSet("SecretID") {
				config.Type = &AppRole{
					RoleID:   viper.GetString("RoleID"),
					SecretID: viper.GetString("SecretID"),
					Client:   client,
				}
			} else {
				return nil, errors.New("RoleID and SecretID for approle authentication cannot be empty")
			}
		case "github":
			if viper.IsSet("GithubToken") {
				config.Type = &Github{
					AccessToken: viper.GetString("GithubToken"),
					Client:      client,
				}
			} else {
				return nil, errors.New("GithubToken for github authentication cannot be empty")
			}
		default:
			return nil, errors.New("Must provide a supported Authentication Type")
		}
	case "secretmanager":
		switch auth {
		case "iam":
			if viper.IsSet("IBMCloudAPIKey") {
				config.Type = &SecretManager{
					IBMCloudAPIKey: viper.GetString("IBMCloudAPIKey"),
					Client:         client,
				}
			} else {
				return nil, errors.New("IBMCloudAPIKey for iam authentication cannot be empty")
			}
		default:
			return nil, errors.New("Must provide a supported Authentication Type")
		}
	default:
		return nil, errors.New("Must provide a supported Vault Type")
	}

	return config, nil
}
