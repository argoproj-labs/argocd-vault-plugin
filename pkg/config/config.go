package config

import (
	"errors"
	"net/http"
	"time"

	"github.com/IBM/argocd-vault-plugin/pkg/auth/ibmsecretmanager"
	"github.com/IBM/argocd-vault-plugin/pkg/auth/vault"
	"github.com/IBM/argocd-vault-plugin/pkg/backends"
	"github.com/IBM/argocd-vault-plugin/pkg/types"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
)

// Config is used to decide the backend and auth type
type Config struct {
	Address    string
	PathPrefix string
	types.Backend
	VaultClient *api.Client
}

// New returns a new Config struct
func New(viper *viper.Viper) (*Config, error) {
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

	var auth types.AuthType
	switch viper.GetString("TYPE") {
	case "vault":
		switch authType {
		case "approle":
			if viper.IsSet("ROLE_ID") && viper.IsSet("SECRET_ID") {
				auth = vault.NewAppRoleAuth(viper.GetString("ROLE_ID"), viper.GetString("SECRET_ID"))
			} else {
				return nil, errors.New("ROLE_ID and SECRET_ID for approle authentication cannot be empty")
			}
		case "github":
			if viper.IsSet("GITHUB_TOKEN") {
				auth = vault.NewGithubAuth(viper.GetString("GITHUB_TOKEN"))
			} else {
				return nil, errors.New("GITHUB_TOKEN for github authentication cannot be empty")
			}
		default:
			return nil, errors.New("Must provide a supported Authentication Type")
		}
		config.Backend = backends.NewVaultBackend(auth, apiClient, kvVersion)
	case "secretmanager":
		switch authType {
		case "iam":
			if viper.IsSet("IBM_API_KEY") {
				auth = ibmsecretmanager.NewIAMAuth(viper.GetString("IBM_API_KEY"))
			} else {
				return nil, errors.New("IBM_API_KEY for iam authentication cannot be empty")
			}
		default:
			return nil, errors.New("Must provide a supported Authentication Type")
		}
		config.Backend = backends.NewIBMSecretManagerBackend(auth, apiClient)
	default:
		return nil, errors.New("Must provide a supported Vault Type")
	}

	return config, nil
}
