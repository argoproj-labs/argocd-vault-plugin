package config

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	gcpsm "cloud.google.com/go/secretmanager/apiv1"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/keyvault/keyvault"
	kvauth "github.com/Azure/azure-sdk-for-go/services/keyvault/auth"
	"github.com/IBM/argocd-vault-plugin/pkg/auth/vault"
	"github.com/IBM/argocd-vault-plugin/pkg/backends"
	"github.com/IBM/argocd-vault-plugin/pkg/kube"
	"github.com/IBM/argocd-vault-plugin/pkg/types"
	"github.com/IBM/go-sdk-core/v5/core"
	ibmsm "github.com/IBM/secrets-manager-go-sdk/secretsmanagerv1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awssm "github.com/aws/aws-sdk-go/service/secretsmanager"
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
			case types.TokenAuth:
				if v.IsSet(api.EnvVaultToken) {
					auth = &vault.TokenAuth{}
				} else {
					return nil, fmt.Errorf("%s for token authentication cannot be empty", api.EnvVaultToken)
				}
			default:
				return nil, errors.New("Must provide a supported Authentication Type")
			}
			backend = backends.NewVaultBackend(auth, apiClient, v.GetString(types.EnvAvpKvVersion))
		}
	case types.IBMSecretsManagerbackend:
		{
			// Get instance URL from IBM specific env variable or fallback to $VAULT_ADDR
			url := v.GetString(types.EnvAvpIBMInstanceURL)
			if !v.IsSet(types.EnvAvpIBMInstanceURL) {
				if !v.IsSet(types.EnvVaultAddress) {
					return nil, fmt.Errorf("%s or %s required for IBM Secrets Manager", types.EnvAvpIBMInstanceURL, types.EnvVaultAddress)
				}
				url = v.GetString(types.EnvVaultAddress)
			}

			client, err := ibmsm.NewSecretsManagerV1(&ibmsm.SecretsManagerV1Options{
				Authenticator: &core.IamAuthenticator{ApiKey: v.GetString(types.EnvAvpIBMAPIKey)},
				URL:           url,
			})
			if err != nil {
				return nil, err
			}

			client.EnableRetries(types.IBMMaxRetries, time.Duration(types.IBMRetryIntervalSeconds)*time.Second)

			backend = backends.NewIBMSecretsManagerBackend(client)
		}
	case types.AWSSecretsManagerbackend:
		{
			if !v.IsSet(types.EnvAWSRegion) { // issue warning when using default region
				log.Printf("Warning: %s env var not set, using AWS region %s.\n", types.EnvAWSRegion, types.AwsDefaultRegion)
				v.Set(types.EnvAWSRegion, types.AwsDefaultRegion)
			}

			s, err := session.NewSession(&aws.Config{
				Region: aws.String(v.GetString(types.EnvAWSRegion)),
			})
			if err != nil {
				return nil, err
			}

			client := awssm.New(s)
			backend = backends.NewAWSSecretsManagerBackend(client)
		}
	case types.GCPSecretManagerbackend:
		{
			ctx := context.Background()
			client, err := gcpsm.NewClient(ctx)
			if err != nil {
				return nil, err
			}
			backend = backends.NewGCPSecretManagerBackend(ctx, client)
		}
	case types.AzureKeyVaultbackend:
		{
			authorizer, err := kvauth.NewAuthorizerFromEnvironment()
			if err != nil {
				return nil, err
			}

			basicClient := keyvault.New()
			basicClient.Authorizer = authorizer
			backend = backends.NewAzureKeyVaultBackend(basicClient)
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

	for k, viperValue := range v.AllSettings() {
		if strings.HasPrefix(k, "vault") || strings.HasPrefix(k, "aws") {
			var value string
			switch viperValue.(type) {
			case bool:
				value = strconv.FormatBool(viperValue.(bool))
			default:
				value = viperValue.(string)
			}
			os.Setenv(strings.ToUpper(k), value)
		}
	}

	return nil
}
