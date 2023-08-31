package backends

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"time"
)

// AzureKeyVault is a struct for working with an Azure Key Vault backend
type AzureKeyVault struct {
	Credential    azcore.TokenCredential
	ClientBuilder func(vaultURL string, credential azcore.TokenCredential, options *azsecrets.ClientOptions) (AzSecretsClient, error)
}

type AzSecretsClient interface {
	GetSecret(ctx context.Context, name string, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error)
	NewListSecretPropertiesPager(options *azsecrets.ListSecretPropertiesOptions) *runtime.Pager[azsecrets.ListSecretPropertiesResponse]
}

// NewAzureKeyVaultBackend initializes a new Azure Key Vault backend
func NewAzureKeyVaultBackend(credential azcore.TokenCredential, clientBuilder func(vaultURL string, credential azcore.TokenCredential, options *azsecrets.ClientOptions) (*azsecrets.Client, error)) *AzureKeyVault {
	return &AzureKeyVault{
		Credential: credential,
		ClientBuilder: func(vaultURL string, credential azcore.TokenCredential, options *azsecrets.ClientOptions) (AzSecretsClient, error) {
			return clientBuilder(vaultURL, credential, options)
		},
	}
}

// Login does nothing as a "login" is handled on the instantiation of the Azure SDK
func (a *AzureKeyVault) Login() error {
	return nil
}

// GetSecrets gets secrets from Azure Key Vault and returns the formatted data
// For Azure Key Vault, `kvpath` is the unique name of your vault
// For Azure use the version here not make really sens as each secret have a different version but let support it
func (a *AzureKeyVault) GetSecrets(kvpath string, version string, _ map[string]string) (map[string]interface{}, error) {
	kvpath = fmt.Sprintf("https://%s.vault.azure.net", kvpath)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	verboseOptionalVersion("Azure Key Vault list all secrets from vault %s", version, kvpath)

	client, err := a.ClientBuilder(kvpath, a.Credential, nil)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})

	pager := client.NewListSecretPropertiesPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, secretVersion := range page.Value {
			// Azure Key Vault has ability to enable/disable a secret, so lets honour that
			if !*secretVersion.Attributes.Enabled {
				continue
			}
			name := secretVersion.ID.Name()
			// Secret version matched given version ?
			if version == "" || secretVersion.ID.Version() == version {
				verboseOptionalVersion("Azure Key Vault getting secret %s from vault %s", version, name, kvpath)
				secret, err := client.GetSecret(ctx, name, version, nil)
				if err != nil {
					return nil, err
				}
				utils.VerboseToStdErr("Azure Key Vault get secret response %v", secret)
				data[name] = *secret.Value
			} else {
				verboseOptionalVersion("Azure Key Vault getting secret %s from vault %s", version, name, kvpath)
				secret, err := client.GetSecret(ctx, name, version, nil)
				if err != nil || !*secretVersion.Attributes.Enabled {
					utils.VerboseToStdErr("Azure Key Vault get versioned secret not found %s", err)
					continue
				}
				utils.VerboseToStdErr("Azure Key Vault get versioned secret response %v", secret)
				data[name] = *secret.Value
			}
		}
	}
	return data, nil
}

// GetIndividualSecret will get the specific secret (placeholder) from the SM backend
// For Azure Key Vault, `kvpath` is the unique name of your vault
// Secrets (placeholders) are directly addressable via the API, so only one call is needed here
func (a *AzureKeyVault) GetIndividualSecret(kvpath, secret, version string, annotations map[string]string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	verboseOptionalVersion("Azure Key Vault getting individual secret %s from vault %s", version, secret, kvpath)

	kvpath = fmt.Sprintf("https://%s.vault.azure.net", kvpath)
	client, err := a.ClientBuilder(kvpath, a.Credential, nil)
	if err != nil {
		return nil, err
	}

	data, err := client.GetSecret(ctx, secret, version, nil)
	if err != nil {
		return nil, err
	}

	utils.VerboseToStdErr("Azure Key Vault get individual secret response %v", data)

	return *data.Value, nil
}

func verboseOptionalVersion(format string, version string, message ...interface{}) {
	if version == "" {
		utils.VerboseToStdErr(format, message...)
	} else {
		utils.VerboseToStdErr(format+" at version %s", append(message, version)...)
	}
}
