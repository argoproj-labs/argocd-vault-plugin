package vault

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/vault/api"
)

// Config TODO
type Config struct {
	Address    string
	PathPrefix string
	Type       VaultType
}

// NewConfig returns a new Config struct
func NewConfig() (*Config, error) {
	config := &Config{
		Address:    getEnv("AVP_VAULT_ADDR", ""),
		PathPrefix: getEnv("AVP_PATH_PREFIX", ""),
	}

	var httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	apiClient, err := api.NewClient(&api.Config{Address: config.Address, HttpClient: httpClient})
	if err != nil {
		return nil, err
	}

	client := &Client{
		VaultAPIClient: apiClient,
	}

	auth := getEnv("AVP_AUTH_TYPE", "")

	switch getEnv("AVP_TYPE", "") {
	case "vault":
		switch auth {
		case "approle":
			config.Type = &AppRole{
				RoleID:   getEnv("AVP_ROLE_ID", ""),
				SecretID: getEnv("AVP_SECRET_ID", ""),
				Client:   client,
			}
		case "github":
			config.Type = &Github{
				AccessToken: getEnv("AVP_GITHUB_TOKEN", ""),
				Client:      client,
			}
		default:
			return nil, errors.New("Must provide a supported Authentication Type")
		}
	case "secretmanager":
		switch auth {
		case "iam":
			config.Type = &SecretManager{
				IBMCloudAPIKey: getEnv("AVP_IBM_API_KEY", ""),
				Client:         client,
			}
		default:
			return nil, errors.New("Must provide a supported Authentication Type")
		}
	default:
		return nil, errors.New("Must provide a supported Vault Type")
	}

	return config, nil
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
