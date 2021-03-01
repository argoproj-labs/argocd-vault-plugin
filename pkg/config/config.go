package config

import (
	"errors"
	"net/http"

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
func New(viper *viper.Viper, httpClient *http.Client) (*Config, error) {

	// Set Defaults
	viper.SetDefault("VAULT_ADDR", "http://127.0.0.1:8200")
	viper.SetDefault("KV_VERSION", "2")

	// Instantiate Env
	viper.SetEnvPrefix("AVP")
	viper.AutomaticEnv()

	config := &Config{
		Address:    viper.GetString("VAULT_ADDR"),
		PathPrefix: viper.GetString("PATH_PREFIX"),
	}

	apiConfig := &api.Config{
		Address:    viper.GetString("VAULT_ADDR"),
		HttpClient: httpClient,
	}

	tlsConfig := &api.TLSConfig{}

	if viper.IsSet("VAULT_CAPATH") {
		tlsConfig.CAPath = viper.GetString("VAULT_CAPATH")
	}

	if viper.IsSet("VAULT_CACERT") {
		tlsConfig.CACert = viper.GetString("VAULT_CACERT")
	}

	if viper.IsSet("VAULT_SKIP_VERIFY") {
		tlsConfig.Insecure = viper.GetBool("VAULT_SKIP_VERIFY")
	}

	if err := apiConfig.ConfigureTLS(tlsConfig); err != nil {
		return nil, err
	}

	apiClient, err := api.NewClient(apiConfig)
	if err != nil {
		return nil, err
	}

	if viper.IsSet("VAULT_NAMESPACE") {
		apiClient.SetNamespace(viper.GetString("VAULT_NAMESPACE"))
	}

	config.VaultClient = apiClient

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
		case "k8s":
			if viper.IsSet("K8S_MOUNT_POINT") && viper.IsSet("K8S_ROLE") {
				tokenPath := ""
				if viper.IsSet("K8S_TOKEN_PATH") {
					tokenPath = viper.GetString("K8S_TOKEN_PATH")
				}
				auth = vault.NewK8sAuth(
					viper.GetString("K8S_MOUNT_POINT"),
					viper.GetString("K8S_ROLE"),
					tokenPath,
				)
			} else {
				return nil, errors.New("K8S_MOUNT_POINT or K8S_ROLE cannot be empty when using Kubernetes Auth")
			}
		default:
			return nil, errors.New("Must provide a supported Authentication Type")
		}
		config.Backend = backends.NewVaultBackend(auth, apiClient, viper.GetString("KV_VERSION"))
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
