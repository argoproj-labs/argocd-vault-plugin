package backends_test

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/backends"
)

type mockClientProxy struct {
	simulateError string
	secretPrefix  string
}

func newMockClientProxy(simulateError, secretPrefix string) *mockClientProxy {
	return &mockClientProxy{
		simulateError: simulateError,
		secretPrefix:  secretPrefix,
	}
}

func makeSecretProperties(id azsecrets.ID, enable bool) *azsecrets.SecretProperties {
	return &azsecrets.SecretProperties{
		ID: &id,
		Attributes: &azsecrets.SecretAttributes{
			Enabled: &enable,
		},
	}
}

func makeResponse(id azsecrets.ID, value string, err error) (azsecrets.GetSecretResponse, error) {
	return azsecrets.GetSecretResponse{
		Secret: azsecrets.Secret{
			ID:    &id,
			Value: &value,
		},
	}, err
}

func (c *mockClientProxy) NewListSecretPropertiesPager(options *azsecrets.ListSecretPropertiesOptions) *runtime.Pager[azsecrets.ListSecretPropertiesResponse] {
	var pageCount = 0
	pager := runtime.NewPager(runtime.PagingHandler[azsecrets.ListSecretPropertiesResponse]{
		More: func(current azsecrets.ListSecretPropertiesResponse) bool {
			return pageCount == 0
		},
		Fetcher: func(ctx context.Context, current *azsecrets.ListSecretPropertiesResponse) (azsecrets.ListSecretPropertiesResponse, error) {
			pageCount++
			var a []*azsecrets.SecretProperties
			if c.simulateError == "fetch_error" {
				return azsecrets.ListSecretPropertiesResponse{}, errors.New("fetch error")
			}
			a = append(a, makeSecretProperties(azsecrets.ID(c.secretPrefix+"simple/v2"), true))
			a = append(a, makeSecretProperties(azsecrets.ID(c.secretPrefix+"second/v2"), true))
			a = append(a, makeSecretProperties(azsecrets.ID(c.secretPrefix+"disabled/v2"), false))
			return azsecrets.ListSecretPropertiesResponse{
				SecretPropertiesListResult: azsecrets.SecretPropertiesListResult{
					Value: a,
				},
			}, nil
		},
	})
	return pager
}

func (c *mockClientProxy) GetSecret(ctx context.Context, name string, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error) {
	if name == "simple" && (version == "" || version == "v1") {
		return makeResponse(azsecrets.ID(c.secretPrefix+"simple/v1"), "a_value_v1", nil)
	} else if name == "simple" && version == "v2" {
		return makeResponse(azsecrets.ID(c.secretPrefix+"simple/v2"), "a_value_v2", nil)
	} else if name == "second" && (version == "" || version == "v2") {
		return makeResponse(azsecrets.ID(c.secretPrefix+"second/v2"), "a_second_value_v2", nil)
	}
	return makeResponse(azsecrets.ID(""), "", errors.New("secret not found"))
}

func newAzureKeyVaultBackendMock(simulateError, secretPrefix string) *backends.AzureKeyVault {
	return &backends.AzureKeyVault{
		Credential: nil,
		ClientBuilder: func(vaultURL string, credential azcore.TokenCredential, options *azsecrets.ClientOptions) (backends.AzSecretsClient, error) {
			return newMockClientProxy(simulateError, secretPrefix), nil
		},
	}
}

func TestAzGetSecrets(t *testing.T) {
	tests := []struct {
		name         string
		secretPrefix string
		cloudEnv     string
	}{
		{"Azure", "https://myvaultname.vault.azure.net/keys/", ""},
		{"AzureChina", "https://myvaultname.vault.azure.cn/keys/", "azurechina"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the environment variable
			os.Setenv("AVP_AZ_CLOUD_NAME", tt.cloudEnv)
			defer os.Unsetenv("AVP_AZ_CLOUD_NAME")

			keyVault := newAzureKeyVaultBackendMock("", tt.secretPrefix)
			res, err := keyVault.GetSecrets("keyvault", "", nil)

			if err != nil {
				t.Fatalf("expected 0 errors but got: %s", err)
			}

			expected := map[string]interface{}{
				"simple": "a_value_v1",
				"second": "a_second_value_v2",
			}
			if !reflect.DeepEqual(res, expected) {
				t.Errorf("expected: %v, got: %v.", expected, res)
			}
		})
	}
}

func TestAzGetSecret(t *testing.T) {
	tests := []struct {
		name         string
		secretPrefix string
		cloudEnv     string
	}{
		{"Azure", "https://myvaultname.vault.azure.net/keys/", ""},
		{"AzureChina", "https://myvaultname.vault.azure.cn/keys/", "azurechina"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the environment variable
			os.Setenv("AVP_AZ_CLOUD_NAME", tt.cloudEnv)
			defer os.Unsetenv("AVP_AZ_CLOUD_NAME")

			keyVault := newAzureKeyVaultBackendMock("", tt.secretPrefix)
			data, err := keyVault.GetIndividualSecret("keyvault", "simple", "", nil)
			if err != nil {
				t.Fatalf("expected 0 errors but got: %s", err)
			}
			expected := "a_value_v1"
			if !reflect.DeepEqual(expected, data) {
				t.Errorf("expected: %s, got: %s.", expected, data)
			}
		})
	}
}