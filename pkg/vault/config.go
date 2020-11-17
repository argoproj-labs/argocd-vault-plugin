package vault

import (
	"errors"
	"os"
)

// Config TODO
type Config struct {
	Address string
	Type    VaultType
}

// NewConfig returns a new Config struct
func NewConfig() (*Config, error) {
	config := &Config{
		Address: getEnv("VAULT_ADDR", ""),
	}

	switch getEnv("VAULT_TYPE", "") {
	case "vault":
		auth := getEnv("AUTH_TYPE", "")
		switch auth {
		case "approle":
			config.Type = &AppRole{
				RoleID:   getEnv("ROLE_ID", ""),
				SecretID: getEnv("SECRET_ID", ""),
			}
		case "github":
			config.Type = &Github{
				AccessToken: getEnv("GITHUB_TOKEN", ""),
			}
		default:
			return nil, errors.New("Must provide a supported Athentication Type")
		}
	case "secretmanager":
		config.Type = &SecretManager{
			IAMToken: getEnv("IAM_TOKEN", ""),
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
