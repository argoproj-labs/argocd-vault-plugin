package backends

import (
	"errors"
	"net/http"
	"time"

	"github.com/IBM/argocd-vault-plugin/pkg/backends/auth"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
)

// Config TODO
type Config struct {
	Address    string
	PathPrefix string
	Backend
	VaultClient *api.Client
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

	config.VaultClient = apiClient

	kvVersion := "2"
	if viper.IsSet("KV_VERSION") {
		kvVersion = viper.GetString("KV_VERSION")
	}

	authType := viper.GetString("AUTH_TYPE")

	switch viper.GetString("TYPE") {
	case "vault":
		vaultBackend := &Vault{
			KvVersion:   kvVersion,
			VaultClient: apiClient,
		}
		switch authType {
		case "approle":
			if viper.IsSet("ROLE_ID") && viper.IsSet("SECRET_ID") {
				vaultBackend.AuthType = &auth.AppRole{
					RoleID:   viper.GetString("ROLE_ID"),
					SecretID: viper.GetString("SECRET_ID"),
				}
			} else {
				return nil, errors.New("ROLE_ID and SECRET_ID for approle authentication cannot be empty")
			}
		case "github":
			if viper.IsSet("GITHUB_TOKEN") {
				vaultBackend.AuthType = &auth.Github{
					AccessToken: viper.GetString("GITHUB_TOKEN"),
				}
			} else {
				return nil, errors.New("GITHUB_TOKEN for github authentication cannot be empty")
			}
		default:
			return nil, errors.New("Must provide a supported Authentication Type")
		}
		config.Backend = vaultBackend
	case "secretmanager":
		switch authType {
		case "iam":
			if viper.IsSet("IBM_API_KEY") {
				config.Backend = &SecretManager{
					IBMCloudAPIKey: viper.GetString("IBM_API_KEY"),
					VaultClient:    apiClient,
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
