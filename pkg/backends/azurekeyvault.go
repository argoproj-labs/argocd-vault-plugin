package backends

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/keyvault/keyvault"
	"github.com/Azure/go-autorest/autorest"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"path"
	"os"
	"strings"
	"time"
)

// AzureKeyVault is a struct for working with an Azure Key Vault backend
type AzureKeyVault struct {
	Client keyvault.BaseClient
}

// authResult contains the subset of results from token acquisition operation in ConfidentialClientApplication
// For details see https://aka.ms/msal-net-authenticationresult
type authResult struct {
	accessToken    string
	expiresOn      time.Time
	grantedScopes  []string
	declinedScopes []string
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

// ClientAssertionBearerAuthorizerCallback gets an Azure AD token for authentication to Azure using an identity
func ClientAssertionBearerAuthorizerCallback(tenantID, resource string) (*autorest.BearerAuthorizer, error) {
	// Azure AD Workload Identity webhook will inject the following env vars
	// 	AZURE_CLIENT_ID with the clientID set in the service account annotation
	// 	AZURE_TENANT_ID with the tenantID set in the service account annotation. If not defined, then
	// 	the tenantID provided via azure-wi-webhook-config for the webhook will be used.
	// 	AZURE_FEDERATED_TOKEN_FILE is the service account token path
	// 	AZURE_AUTHORITY_HOST is the AAD authority hostname
	clientID := os.Getenv("AZURE_CLIENT_ID")
	tokenFilePath := os.Getenv("AZURE_FEDERATED_TOKEN_FILE")
	authorityHost := os.Getenv("AZURE_AUTHORITY_HOST")

	utils.VerboseToStdErr("Azure Client ID %v", clientID)
	utils.VerboseToStdErr("Azure TokenFilePath %v", tokenFilePath)
	utils.VerboseToStdErr("Azure authorityHost %v", authorityHost)
	// generate a token using the msal confidential client
	// this will always generate a new token request to AAD

	cred := confidential.NewCredFromAssertionCallback(func(context.Context, confidential.AssertionRequestOptions) (string, error) {
		return ReadJWTFromFS(tokenFilePath)
	})
	utils.VerboseToStdErr("cred %v", cred)
	// create the confidential client to request an AAD token
	confidentialClientApp, err := confidential.New(
		clientID,
		cred,
		confidential.WithAuthority(fmt.Sprintf("%s%s/oauth2/token", authorityHost, tenantID)))
	if err != nil {
		utils.VerboseToStdErr("Error in confidental new")
		return nil, err
	}

	// trim the suffix / if exists
	resource = strings.TrimSuffix(resource, "/")
	// .default needs to be added to the scope
	if !strings.HasSuffix(resource, ".default") {
		resource += "/.default"
	}
	utils.VerboseToStdErr("resource %v", resource)

	result, err := confidentialClientApp.AcquireTokenByCredential(context.Background(), []string{resource})
	if err != nil {
		return nil, err
	}
	utils.VerboseToStdErr("access token %v", result.AccessToken)

	return autorest.NewBearerAuthorizer(authResult{
		accessToken:    result.AccessToken,
		expiresOn:      result.ExpiresOn,
		grantedScopes:  result.GrantedScopes,
		declinedScopes: result.DeclinedScopes,
	}), nil
}

// OAuthToken implements the OAuthTokenProvider interface.  It returns the current access token.
func (ar authResult) OAuthToken() string {
	return ar.accessToken
}

// ReadJWTFromFS reads the jwt from file system
func ReadJWTFromFS(tokenFilePath string) (string, error) {
	token, err := os.ReadFile(tokenFilePath)
	if err != nil {
		return "", err
	}
	return string(token), nil
}
