package backends

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/keyvault/keyvault"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"path"
	"strings"
	"time"
)

// AzureKeyVault is a struct for working with an Azure Key Vault backend
type AzureKeyVault struct {
	Client keyvault.BaseClient
}

// NewAzureKeyVaultBackend initializes a new Azure Key Vault backend
func NewAzureKeyVaultBackend(client keyvault.BaseClient) *AzureKeyVault {
	return &AzureKeyVault{
		Client: client,
	}
}

// Login does nothing as a "login" is handled on the instantiation of the Azure SDK
func (a *AzureKeyVault) Login() error {
	return nil
}

// GetSecrets gets secrets from Azure Key Vault and returns the formatted data
// For Azure Key Vault, `kvpath` is the unique name of your vault
func (a *AzureKeyVault) GetSecrets(kvpath string, version string, _ map[string]string) (map[string]interface{}, error) {
	kvpath = fmt.Sprintf("https://%s.vault.azure.net", kvpath)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	data := make(map[string]interface{})

	utils.VerboseToStdErr("Azure Key Vault listing secrets in vault %v", kvpath)
	secretList, err := a.Client.GetSecretsComplete(ctx, kvpath, nil)
	if err != nil {
		return nil, err
	}

	utils.VerboseToStdErr("Azure Key Vault list secrets response %v", secretList)
	// Gather all secrets in Key Vault

	for ; secretList.NotDone(); secretList.NextWithContext(ctx) {
		secret := path.Base(*secretList.Value().ID)
		if version == "" {
			utils.VerboseToStdErr("Azure Key Vault getting secret %s from vault %s", secret, kvpath)
			secretResp, err := a.Client.GetSecret(ctx, kvpath, secret, "")
			if err != nil {
				return nil, err
			}

			utils.VerboseToStdErr("Azure Key Vault get unversioned secret response %v", secretResp)
			data[secret] = *secretResp.Value
			continue
		}
		// In Azure Key Vault the versions of a secret is first shown after running GetSecretVersions. So we need
		// to loop through the versions for each secret in order to find the secret that has the specific version.
		secretVersions, _ := a.Client.GetSecretVersionsComplete(ctx, kvpath, secret, nil)
		for ; secretVersions.NotDone(); secretVersions.NextWithContext(ctx) {
			secretVersion := secretVersions.Value()
			// Azure Key Vault has ability to enable/disable a secret, so lets honour that
			if !*secretVersion.Attributes.Enabled {
				continue
			}
			// Secret version matched given version
			if strings.Contains(*secretVersion.ID, version) {
				utils.VerboseToStdErr("Azure Key Vault getting secret %s from vault %s at version %s", secret, kvpath, version)
				secretResp, err := a.Client.GetSecret(ctx, kvpath, secret, version)
				if err != nil {
					return nil, err
				}

				utils.VerboseToStdErr("Azure Key Vault get versioned secret response %v", secretResp)
				data[secret] = *secretResp.Value
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

	utils.VerboseToStdErr("Azure Key Vault getting secret %s from vault %s at version %s", secret, kvpath, version)

	kvpath = fmt.Sprintf("https://%s.vault.azure.net", kvpath)
	data, err := a.Client.GetSecret(ctx, kvpath, secret, version)
	if err != nil {
		return nil, err
	}

	utils.VerboseToStdErr("Azure Key Vault get versioned secret response %v", data)

	return *data.Value, nil
}
