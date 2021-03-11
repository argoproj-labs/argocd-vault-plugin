package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/IBM/argocd-vault-plugin/pkg/auth/ibmsecretmanager"
	"github.com/IBM/argocd-vault-plugin/pkg/auth/vault"
	"github.com/IBM/argocd-vault-plugin/pkg/backends"
	"github.com/IBM/argocd-vault-plugin/pkg/kube"
	"github.com/IBM/argocd-vault-plugin/pkg/types"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
)

const (
	// Environment Variable Constants
	envAvpType         = "AVP_TYPE"
	envAvpRoleID       = "AVP_ROLE_ID"
	envAvpSecretID     = "AVP_SECRET_ID"
	envAvpAuthType     = "AVP_AUTH_TYPE"
	envAvpGithubToken  = "AVP_GITHUB_TOKEN"
	envAvpK8sRole      = "AVP_K8S_ROLE"
	envAvpK8sMountPath = "AVP_K8S_MOUNT_PATH"
	envAvpK8sTokenPath = "AVP_K8S_TOKEN_PATH"
	envAvpIbmAPIKey    = "AVP_IBM_API_KEY"
	envAvpKvVersion    = "AVP_KV_VERSION"
	envAvpPathPrefix   = "AVP_PATH_PREFIX"

	// Backend and Auth Constants
	vaultBackend            = "vault"
	ibmSecretManagerbackend = "ibmsecretmanager"
	k8sAuth                 = "k8s"
	approleAuth             = "approle"
	githubAuth              = "github"
	iamAuth                 = "iam"
)

// Options TODO
type Options struct {
	SecretName string
	ConfigPath string
}

// Config is used to decide the backend and auth type
type Config struct {
	Backend     types.Backend
	VaultClient *api.Client
	PathPrefix  string
}

// New returns a new Config struct
func New(co *Options) (*Config, error) {
	viper := viper.New()

	// Set Defaults
	viper.SetDefault(envAvpKvVersion, "2")

	// Read in config file or kubernetes secret and set as env vars
	err := readConfigOrSecret(co.SecretName, co.ConfigPath)
	if err != nil {
		return nil, err
	}

	// Instantiate Env
	viper.AutomaticEnv()

	apiClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, err
	}

	authType := viper.GetString(envAvpAuthType)

	var auth types.AuthType
	var backend types.Backend

	switch viper.GetString(envAvpType) {
	case vaultBackend:
		switch authType {
		case approleAuth:
			if viper.IsSet(envAvpRoleID) && viper.IsSet(envAvpSecretID) {
				auth = vault.NewAppRoleAuth(viper.GetString(envAvpRoleID), viper.GetString(envAvpSecretID))
			} else {
				return nil, fmt.Errorf("%s and %s for approle authentication cannot be empty", envAvpRoleID, envAvpSecretID)
			}
		case githubAuth:
			if viper.IsSet(envAvpGithubToken) {
				auth = vault.NewGithubAuth(viper.GetString(envAvpGithubToken))
			} else {
				return nil, fmt.Errorf("%s for github authentication cannot be empty", envAvpGithubToken)
			}
		case k8sAuth:
			if viper.IsSet(envAvpK8sRole) {
				auth = vault.NewK8sAuth(
					viper.GetString(envAvpK8sRole),
					viper.GetString(envAvpK8sMountPath),
					viper.GetString(envAvpK8sTokenPath),
				)
			} else {
				return nil, fmt.Errorf("%s cannot be empty when using Kubernetes Auth", envAvpK8sRole)
			}
		default:
			return nil, errors.New("Must provide a supported Authentication Type")
		}
		backend = backends.NewVaultBackend(auth, apiClient, viper.GetString(envAvpKvVersion))
	case ibmSecretManagerbackend:
		switch authType {
		case iamAuth:
			if viper.IsSet(envAvpIbmAPIKey) {
				auth = ibmsecretmanager.NewIAMAuth(viper.GetString(envAvpIbmAPIKey))
			} else {
				return nil, fmt.Errorf("%s for iam authentication cannot be empty", envAvpIbmAPIKey)
			}
		default:
			return nil, errors.New("Must provide a supported Authentication Type")
		}
		backend = backends.NewIBMSecretManagerBackend(auth, apiClient)
	default:
		return nil, errors.New("Must provide a supported Vault Type")
	}

	return &Config{
		Backend:     backend,
		VaultClient: apiClient,
		PathPrefix:  viper.GetString(envAvpPathPrefix),
	}, nil
}

func readConfigOrSecret(secretName, configPath string) error {
	// If a secret name is passed, pull config from Kubernetes
	if secretName != "" {
		localClient, err := kube.NewClient()
		if err != nil {
			return err
		}
		yaml, err := localClient.ReadSecret(secretName)
		if err != nil {
			return err
		}
		viper.SetConfigType("yaml")
		viper.ReadConfig(bytes.NewBuffer(yaml))
	}

	// If a config file path is passed, read in that file and overwrite all other
	if configPath != "" {
		viper.SetConfigFile(configPath)
		err := viper.ReadInConfig()
		if err != nil {
			return err
		}
	}

	for k, v := range viper.AllSettings() {
		if strings.HasPrefix(k, "vault") {
			var value string
			switch v.(type) {
			case bool:
				value = strconv.FormatBool(v.(bool))
			default:
				value = v.(string)
			}
			os.Setenv(strings.ToUpper(k), value)
		}
	}

	return nil
}
