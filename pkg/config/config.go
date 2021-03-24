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

// const (
// 	// Environment Variable Constants
// 	envAvpType         = "AVP_TYPE"
// 	envAvpRoleID       = "AVP_ROLE_ID"
// 	envAvpSecretID     = "AVP_SECRET_ID"
// 	envAvpAuthType     = "AVP_AUTH_TYPE"
// 	envAvpGithubToken  = "AVP_GITHUB_TOKEN"
// 	envAvpK8sRole      = "AVP_K8S_ROLE"
// 	envAvpK8sMountPath = "AVP_K8S_MOUNT_PATH"
// 	envAvpK8sTokenPath = "AVP_K8S_TOKEN_PATH"
// 	envAvpIbmAPIKey    = "AVP_IBM_API_KEY"
// 	EnvAvpKvVersion    = "AVP_KV_VERSION"
// 	envAvpPathPrefix   = "AVP_PATH_PREFIX"
//
// 	// Backend and Auth Constants
// 	vaultBackend            = "vault"
// 	ibmSecretManagerbackend = "ibmsecretmanager"
// 	k8sAuth                 = "k8s"
// 	approleAuth             = "approle"
// 	githubAuth              = "github"
// 	iamAuth                 = "iam"
// )

// Options TODO
type Options struct {
	SecretName string
	ConfigPath string
}

// Config is used to decide the backend and auth type
type Config struct {
	Backend     types.Backend
	VaultClient *api.Client
}

// New returns a new Config struct
func New(v *viper.Viper, co *Options) (*Config, error) {

	// Set Defaults
	v.SetDefault(types.EnvAvpKvVersion, "2")

	// Read in config file or kubernetes secret and set as env vars
	err := readConfigOrSecret(co.SecretName, co.ConfigPath, v)
	if err != nil {
		return nil, err
	}

	// Instantiate Env
	v.AutomaticEnv()

	apiClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, err
	}

	authType := v.GetString(types.EnvAvpAuthType)

	var auth types.AuthType
	var backend types.Backend

	switch v.GetString(types.EnvAvpType) {
	case types.VaultBackend:
		switch authType {
		case types.ApproleAuth:
			if v.IsSet(types.EnvAvpRoleID) && v.IsSet(types.EnvAvpSecretID) {
				auth = vault.NewAppRoleAuth(v.GetString(types.EnvAvpRoleID), v.GetString(types.EnvAvpSecretID))
			} else {
				return nil, fmt.Errorf("%s and %s for approle authentication cannot be empty", types.EnvAvpRoleID, types.EnvAvpSecretID)
			}
		case types.GithubAuth:
			if v.IsSet(types.EnvAvpGithubToken) {
				auth = vault.NewGithubAuth(v.GetString(types.EnvAvpGithubToken))
			} else {
				return nil, fmt.Errorf("%s for github authentication cannot be empty", types.EnvAvpGithubToken)
			}
		case types.K8sAuth:
			if v.IsSet(types.EnvAvpK8sRole) {
				auth = vault.NewK8sAuth(
					v.GetString(types.EnvAvpK8sRole),
					v.GetString(types.EnvAvpK8sMountPath),
					v.GetString(types.EnvAvpK8sTokenPath),
				)
			} else {
				return nil, fmt.Errorf("%s cannot be empty when using Kubernetes Auth", types.EnvAvpK8sRole)
			}
		default:
			return nil, errors.New("Must provide a supported Authentication Type")
		}
		backend = backends.NewVaultBackend(auth, apiClient, v.GetString(types.EnvAvpKvVersion))
	case types.IbmSecretManagerbackend:
		switch authType {
		case types.IamAuth:
			if v.IsSet(types.EnvAvpIbmAPIKey) {
				auth = ibmsecretmanager.NewIAMAuth(v.GetString(types.EnvAvpIbmAPIKey))
			} else {
				return nil, fmt.Errorf("%s for iam authentication cannot be empty", types.EnvAvpIbmAPIKey)
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
	}, nil
}

func readConfigOrSecret(secretName, configPath string, v *viper.Viper) error {
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
		v.SetConfigType("yaml")
		v.ReadConfig(bytes.NewBuffer(yaml))
	}

	// If a config file path is passed, read in that file and overwrite all other
	if configPath != "" {
		v.SetConfigFile(configPath)
		err := v.ReadInConfig()
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
