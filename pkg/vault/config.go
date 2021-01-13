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
	*Client
}

// NewConfig returns a new Config struct
func NewConfig(viper *viper.Viper) (*Config, error) {
	viper.SetEnvPrefix("AVP")
	viper.AutomaticEnv()

	config := &Config{
		Address:    viper.GetString("VAULT_ADDR"),
		PathPrefix: viper.GetString("PATH_PREFIX"),
	}

	var httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	apiClient, err := api.NewClient(&api.Config{Address: viper.GetString("VAULT_ADDR"), HttpClient: httpClient})
	if err != nil {
		return nil, err
	}

	client := &Client{
		VaultAPIClient: apiClient,
	}
	config.Client = client

	auth := viper.GetString("AUTH_TYPE")

	switch viper.GetString("TYPE") {
	case "vault":
		switch auth {
		case "approle":
			if viper.IsSet("ROLE_ID") && viper.IsSet("SECRET_ID") {
				config.Type = &AppRole{
					RoleID:   viper.GetString("ROLE_ID"),
					SecretID: viper.GetString("SECRET_ID"),
					Client:   client,
				}
			} else {
				return nil, errors.New("ROLE_ID and SECRET_ID for approle authentication cannot be empty")
			}
		case "github":
			if viper.IsSet("GITHUB_TOKEN") {
				config.Type = &Github{
					AccessToken: viper.GetString("GITHUB_TOKEN"),
					Client:      client,
				}
			} else {
				return nil, errors.New("GITHUB_TOKEN for github authentication cannot be empty")
			}
		default:
			return nil, errors.New("Must provide a supported Authentication Type")
		}
	case "secretmanager":
		switch auth {
		case "iam":
			if viper.IsSet("IBM_API_KEY") {
				config.Type = &SecretManager{
					IBMCloudAPIKey: viper.GetString("IBM_API_KEY"),
					Client:         client,
				}
			} else {
				return nil, errors.New("IBM_API_KEY for iam authentication cannot be empty")
			}
		default:
			return nil, errors.New("Must provide a supported Authentication Type")
		}
	default:
		return nil, errors.New("Must provide a supported Vault Type")
	}

	return config, nil
}
