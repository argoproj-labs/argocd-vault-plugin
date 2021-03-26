package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/IBM/argocd-vault-plugin/pkg/auth/ibmsecretsmanager"
	"github.com/IBM/argocd-vault-plugin/pkg/auth/vault"
	"github.com/IBM/argocd-vault-plugin/pkg/backends"
	"github.com/IBM/argocd-vault-plugin/pkg/kube"
	"github.com/IBM/argocd-vault-plugin/pkg/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
)

// Options options that can be passed to a Config struct
type Options struct {
	SecretName string
	ConfigPath string
}

// Config is used to decide the backend and auth type
type Config struct {
	Backend types.Backend
}

// New returns a new Config struct
func New(v *viper.Viper, co *Options) (*Config, error) {

	// Set Defaults
	v.SetDefault(types.EnvAvpKvVersion, "2")
	v.SetDefault("AVP_AWS_REGION", "us-east-2")

	// Read in config file or kubernetes secret and set as env vars
	err := readConfigOrSecret(co.SecretName, co.ConfigPath, v)
	if err != nil {
		return nil, err
	}

	// Instantiate Env
	v.AutomaticEnv()

	authType := v.GetString(types.EnvAvpAuthType)

	var auth types.AuthType
	var backend types.Backend

	switch v.GetString(types.EnvAvpType) {
	case types.VaultBackend:
		{
			apiClient, err := api.NewClient(api.DefaultConfig())
			if err != nil {
				return nil, err
			}

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
		}
	case types.IBMSecretsManagerbackend:
		{
			apiClient, err := api.NewClient(api.DefaultConfig())
			if err != nil {
				return nil, err
			}

			switch authType {
			case types.IAMAuth:
				if v.IsSet(types.EnvAvpIBMAPIKey) {
					auth = ibmsecretsmanager.NewIAMAuth(v.GetString(types.EnvAvpIBMAPIKey))
				} else {
					return nil, fmt.Errorf("%s for iam authentication cannot be empty", types.EnvAvpIBMAPIKey)
				}
			default:
				return nil, errors.New("Must provide a supported Authentication Type")
			}
			backend = backends.NewIBMSecretsManagerBackend(auth, apiClient)
		}
	case types.AWSSecretsManagerbackend:
		{
			if !v.IsSet(types.EnvAWSAccessKey) || !v.IsSet(types.EnvAWSSecretAccessKey) {
				return nil, fmt.Errorf("Must provide %s and %s for backend type %s",
					types.EnvAWSAccessKey,
					types.EnvAWSSecretAccessKey,
					types.AWSSecretsManagerbackend,
				)
			}
			s := session.Must(session.NewSession(&aws.Config{
				Region: aws.String(v.GetString(types.EnvAWSRegion)),
			}))
			client := secretsmanager.New(s)
			backend = backends.NewAWSSecretsManagerBackend(client)
		}
	default:
		return nil, errors.New("Must provide a supported Vault Type")
	}

	return &Config{
		Backend: backend,
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
