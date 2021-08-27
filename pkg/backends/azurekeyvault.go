package backends

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/keyvault/keyvault"
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

	secretList, err := a.Client.GetSecrets(ctx, kvpath, nil)
	if err != nil {
		return nil, err
	}
	// Gather all secrets in Key Vault
	for ; secretList.NotDone(); secretList.NextWithContext(ctx) {
		for _, secretItem := range secretList.Values() {
			secret := path.Base(*secretItem.ID)
			if version == "" {
				secretResp, err := a.Client.GetSecret(ctx, kvpath, secret, "")
				if err != nil {
					return nil, err
				}
				data[secret] = *secretResp.Value
				continue
			}
			// In Azure Key Vault the versions of a secret is first shown after running GetSecretVersions. So we need
			// to loop through the versions for each secret in order to find the secret that has the specific version.
			secretVersions, _ := a.Client.GetSecretVersions(ctx, kvpath, secret, nil)
			for ; secretVersions.NotDone(); secretVersions.NextWithContext(ctx) {
				for _, secretVersion := range secretVersions.Values() {
					// Azure Key Vault has ability to enable/disable a secret, so lets honour that
					if !*secretVersion.Attributes.Enabled {
						continue
					}
					// Secret version matched given version
					if strings.Contains(*secretVersion.ID, version) {
						secretResp, err := a.Client.GetSecret(ctx, kvpath, secret, version)
						if err != nil {
							return nil, err
						}
						data[secret] = *secretResp.Value
					}
				}
			}
		}
	}

	return data, nil
}
