package config

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	gcpsm "cloud.google.com/go/secretmanager/apiv1"
	"github.com/1Password/connect-sdk-go/connect"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/keyvault/keyvault"
	kvauth "github.com/Azure/azure-sdk-for-go/services/keyvault/auth"
	delineasecretserver "github.com/DelineaXPM/tss-sdk-go/v2/server"
	"github.com/IBM/go-sdk-core/v5/core"
	ibmsm "github.com/IBM/secrets-manager-go-sdk/secretsmanagerv2"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/auth/vault"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/backends"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/kube"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/types"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"github.com/aws/aws-sdk-go-v2/config"
	awssm "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hashicorp/vault/api"
	ksm "github.com/keeper-security/secrets-manager-go/core"
	"github.com/spf13/viper"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/iamkey"
	sops "go.mozilla.org/sops/v3/decrypt"
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

var backendPrefixes []string = []string{
	"vault",
	"aws",
	"azure",
	"google",
	"sops",
	"op_connect",
	"k8s_secret",
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
	utils.VerboseToStdErr("reading configuration from environment, overriding any previous settings")
	v.AutomaticEnv()

	utils.VerboseToStdErr("AVP configured with the following settings:\n")
	for k, viperValue := range v.AllSettings() {
		utils.VerboseToStdErr("%s: %s\n", k, viperValue)
	}

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
					auth = vault.NewAppRoleAuth(v.GetString(types.EnvAvpRoleID), v.GetString(types.EnvAvpSecretID), v.GetString(types.EnvAvpMountPath))
				} else {
					return nil, fmt.Errorf("%s and %s for approle authentication cannot be empty", types.EnvAvpRoleID, types.EnvAvpSecretID)
				}
			case types.GithubAuth:
				if v.IsSet(types.EnvAvpGithubToken) {
					auth = vault.NewGithubAuth(v.GetString(types.EnvAvpGithubToken), v.GetString(types.EnvAvpMountPath))
				} else {
					return nil, fmt.Errorf("%s for github authentication cannot be empty", types.EnvAvpGithubToken)
				}
			case types.K8sAuth:
				if v.IsSet(types.EnvAvpK8sRole) {
					// Prefer the K8s mount path if set (for backwards compatibility), use the generic Vault mount path otherwise
					if v.IsSet(types.EnvAvpK8sMountPath) {
						auth = vault.NewK8sAuth(
							v.GetString(types.EnvAvpK8sRole),
							v.GetString(types.EnvAvpK8sMountPath),
							v.GetString(types.EnvAvpK8sTokenPath),
						)
					} else {
						auth = vault.NewK8sAuth(
							v.GetString(types.EnvAvpK8sRole),
							v.GetString(types.EnvAvpMountPath),
							v.GetString(types.EnvAvpK8sTokenPath),
						)
					}
				} else {
					return nil, fmt.Errorf("%s cannot be empty when using Kubernetes Auth", types.EnvAvpK8sRole)
				}
			case types.TokenAuth:
				if v.IsSet(api.EnvVaultToken) {
					auth = &vault.TokenAuth{}
				} else {
					return nil, fmt.Errorf("%s for token authentication cannot be empty", api.EnvVaultToken)
				}
			case types.UserPass:
				if v.IsSet(types.EnvAvpUsername) && v.IsSet(types.EnvAvpPassword) {
					auth = vault.NewUserPassAuth(v.GetString(types.EnvAvpUsername), v.GetString(types.EnvAvpPassword), v.GetString(types.EnvAvpMountPath))
				} else {
					return nil, fmt.Errorf("%s and %s for userpass authentication cannot be empty", types.EnvAvpUsername, types.EnvAvpPassword)
				}
			default:
				return nil, fmt.Errorf("Must provide a supported Authentication Type, received %s", authType)
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

				utils.VerboseToStdErr("falling back to %s in place of %s", types.EnvVaultAddress, types.EnvAvpIBMInstanceURL)
				url = v.GetString(types.EnvVaultAddress)
			}

			client, err := ibmsm.NewSecretsManagerV2(&ibmsm.SecretsManagerV2Options{
				Authenticator: &core.IamAuthenticator{ApiKey: v.GetString(types.EnvAvpIBMAPIKey)},
				URL:           url,
			})
			if err != nil {
				return nil, err
			}

			utils.VerboseToStdErr("IBM Cloud Secrets Manager enabling %d API call retries with %d seconds between tries", types.IBMMaxRetries, types.IBMRetryIntervalSeconds)
			client.EnableRetries(types.IBMMaxRetries, time.Duration(types.IBMRetryIntervalSeconds)*time.Second)

			backend = backends.NewIBMSecretsManagerBackend(client)
		}
	case types.AWSSecretsManagerbackend:
		{
			if !v.IsSet(types.EnvAWSRegion) {
				utils.VerboseToStdErr("warning: %s env var not set, using AWS region %s", types.EnvAWSRegion, types.AwsDefaultRegion)
				v.Set(types.EnvAWSRegion, types.AwsDefaultRegion)
			}

			s, err := config.LoadDefaultConfig(context.TODO(),
				config.WithRegion(v.GetString(types.EnvAWSRegion)),
			)
			if err != nil {
				return nil, err
			}

			client := awssm.NewFromConfig(s)
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
	case types.Sopsbackend:
		{
			backend = backends.NewLocalSecretManagerBackend(sops.File)
		}
	case types.YandexCloudLockboxbackend:
		{
			if !v.IsSet(types.EnvYCLKeyID) ||
				!v.IsSet(types.EnvYCLServiceAccountID) ||
				!v.IsSet(types.EnvYCLPrivateKey) {
				return nil, fmt.Errorf(
					"%s, %s and %s are required for yandex cloud lockbox",
					types.EnvYCLKeyID,
					types.EnvYCLServiceAccountID,
					types.EnvYCLPrivateKey,
				)
			}

			creds, err := ycsdk.ServiceAccountKey(&iamkey.Key{
				Id:         v.GetString(types.EnvYCLKeyID),
				Subject:    &iamkey.Key_ServiceAccountId{ServiceAccountId: v.GetString(types.EnvYCLServiceAccountID)},
				PrivateKey: v.GetString(types.EnvYCLPrivateKey),
			})
			if err != nil {
				return nil, err
			}

			sdk, err := ycsdk.Build(context.Background(), ycsdk.Config{Credentials: creds})
			if err != nil {
				return nil, err
			}

			backend = backends.NewYandexCloudLockboxBackend(sdk.LockboxPayload().Payload())
		}
	case types.OnePasswordConnect:
		{
			client, err := connect.NewClientFromEnvironment()
			if err != nil {
				return nil, err
			}

			backend = backends.NewOnePasswordConnectBackend(client)
		}
	case types.KeeperSecretsManagerBackend:
		{
			if !v.IsSet(types.EnvAvpKSMConfigPath) {
				return nil, fmt.Errorf("%s is required for Keeper Secrets Manager", types.EnvAvpKSMConfigPath)
			}

			client := ksm.NewSecretsManager(&ksm.ClientOptions{
				Config: ksm.NewFileKeyValueStorage(
					v.GetString(types.EnvAvpKSMConfigPath),
				),
			})

			backend = backends.NewKeeperSecretsManagerBackend(client)
		}
	case types.DelineaSecretServerbackend:
		{
			// Get required Delinea specific env variables
			if !v.IsSet(types.EnvAvpDelineaURL) ||
				!v.IsSet(types.EnvAvpDelineaUser) ||
				!v.IsSet(types.EnvAvpDelineaPassword) {
				return nil, fmt.Errorf("%s, %s and %s are required for Delinea Secret Server",
					types.EnvAvpDelineaURL,
					types.EnvAvpDelineaUser,
					types.EnvAvpDelineaPassword,
				)
			}

			tss, _ := delineasecretserver.New(delineasecretserver.Configuration{
				Credentials: delineasecretserver.UserCredential{
					Username: v.GetString(types.EnvAvpDelineaUser),
					Password: v.GetString(types.EnvAvpDelineaPassword),
					Domain:   v.GetString(types.EnvAvpDelineaDomain),
				},
				ServerURL: v.GetString(types.EnvAvpDelineaURL),
			})
			if err != nil {
				return nil, err
			}
			backend = backends.NewDelineaSecretServerBackend(tss)
		}
	case types.KubernetesSecretBackend:
		{
			backend = backends.NewKubernetesSecret()
		}
	default:
		return nil, fmt.Errorf("Must provide a supported Vault Type, received %s", v.GetString(types.EnvAvpType))
	}

	return &Config{
		Backend: backend,
	}, nil
}

func readConfigOrSecret(secretName, configPath string, v *viper.Viper) error {
	// If a secret name is passed, pull config from Kubernetes
	if secretName != "" {
		utils.VerboseToStdErr("reading configuration from secret %s", secretName)

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
		utils.VerboseToStdErr("reading configuration from config file %s, overriding any previous settings", configPath)

		v.SetConfigFile(configPath)
		err := v.ReadInConfig()
		if err != nil {
			return err
		}
	}

	// Check for ArgoCD 2.4 prefixed environment variables
	for _, envVar := range os.Environ() {
		if strings.HasPrefix(envVar, types.EnvArgoCDPrefix) {
			envVarPair := strings.SplitN(envVar, "=", 2)
			key := strings.TrimPrefix(envVarPair[0], types.EnvArgoCDPrefix+"_")
			val := envVarPair[1]
			v.Set(key, val)
		}
	}

	for k, viperValue := range v.AllSettings() {
		for _, prefix := range backendPrefixes {
			if strings.HasPrefix(k, prefix) {
				var value string
				switch viperValue.(type) {
				case bool:
					value = strconv.FormatBool(viperValue.(bool))
				default:
					value = viperValue.(string)
				}
				os.Setenv(strings.ToUpper(k), value)
				utils.VerboseToStdErr("Setting %s to %s for backend SDK", strings.ToUpper(k), value)
			}
		}
	}

	return nil
}
