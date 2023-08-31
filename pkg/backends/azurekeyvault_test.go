package backends_test

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/backends"
	"reflect"
	"testing"
)

const secretNamePrefix = "https://myvaultname.vault.azure.net/keys/"

type mockClientProxy struct {
	simulateError string
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

func newAzureKeyVaultBackendMock(simulateError string) *backends.AzureKeyVault {
	return &backends.AzureKeyVault{
		Credential: nil,
		ClientBuilder: func(vaultURL string, credential azcore.TokenCredential, options *azsecrets.ClientOptions) (backends.AzSecretsClient, error) {
			return &mockClientProxy{
				simulateError: simulateError,
			}, nil
		},
	}
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
			} else if c.simulateError == "get_secret_error" {
				a = append(a, makeSecretProperties(secretNamePrefix+"invalid/v2", true))
			}
			a = append(a, makeSecretProperties(secretNamePrefix+"simple/v2", true))
			a = append(a, makeSecretProperties(secretNamePrefix+"second/v2", true))
			a = append(a, makeSecretProperties(secretNamePrefix+"disabled/v2", false))
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
		return makeResponse(secretNamePrefix+"simple/v1", "a_value_v1", nil)
	} else if name == "simple" && version == "v2" {
		return makeResponse(secretNamePrefix+"simple/v2", "a_value_v2", nil)
	} else if name == "second" && (version == "" || version == "v2") {
		return makeResponse(secretNamePrefix+"second/v2", "a_second_value_v2", nil)
	}
	return makeResponse("", "", errors.New("secret not found"))
}

func TestAzLogin(t *testing.T) {
	var keyVault = newAzureKeyVaultBackendMock("")
	var err = keyVault.Login()
	if err != nil {
		t.Fatalf("expected 0 errors but got: %s", err)
	}
}

func TestAzGetSecret(t *testing.T) {
	var keyVault = newAzureKeyVaultBackendMock("")
	var data, err = keyVault.GetIndividualSecret("keyvault", "simple", "", nil)
	if err != nil {
		t.Fatalf("expected 0 errors but got: %s", err)
	}
	expected := "a_value_v1"
	if !reflect.DeepEqual(expected, data) {
		t.Errorf("expected: %s, got: %s.", expected, data)
	}
}

func TestAzGetSecretWithVersion(t *testing.T) {
	var keyVault = newAzureKeyVaultBackendMock("")
	var data, err = keyVault.GetIndividualSecret("keyvault", "simple", "v2", nil)
	if err != nil {
		t.Fatalf("expected 0 errors but got: %s", err)
	}
	expected := "a_value_v2"
	if !reflect.DeepEqual(expected, data) {
		t.Errorf("expected: %s, got: %s.", expected, data)
	}
}

func TestAzGetSecretWithWrongVersion(t *testing.T) {
	var keyVault = newAzureKeyVaultBackendMock("")
	var _, err = keyVault.GetIndividualSecret("keyvault", "simple", "v3", nil)
	if err == nil {
		t.Fatalf("expected 1 errors but got nil")
	}
	expected := errors.New("secret not found")
	if !reflect.DeepEqual(err, expected) {
		t.Errorf("expected err: %s, got: %s.", expected, err)
	}
}

func TestAzGetSecretNotExist(t *testing.T) {
	var keyVault = newAzureKeyVaultBackendMock("")
	var _, err = keyVault.GetIndividualSecret("keyvault", "not_existing", "", nil)
	if err == nil {
		t.Fatalf("expected 1 errors but got nil")
	}
	expected := errors.New("secret not found")
	if !reflect.DeepEqual(err, expected) {
		t.Errorf("expected err: %s, got: %s.", expected, err)
	}
}

func TestAzGetSecretBuilderError(t *testing.T) {
	var keyVault = &backends.AzureKeyVault{
		Credential: nil,
		ClientBuilder: func(vaultURL string, credential azcore.TokenCredential, options *azsecrets.ClientOptions) (backends.AzSecretsClient, error) {
			return nil, errors.New("boom")
		},
	}
	var _, err = keyVault.GetIndividualSecret("keyvault", "not_existing", "", nil)
	if err == nil {
		t.Fatalf("expected 1 errors but got nil")
	}
	expected := errors.New("boom")
	if !reflect.DeepEqual(err, expected) {
		t.Errorf("expected err: %s, got: %s.", expected, err)
	}
}

func TestAzGetSecrets(t *testing.T) {
	var keyVault = newAzureKeyVaultBackendMock("")
	var res, err = keyVault.GetSecrets("keyvault", "", nil)

	if err != nil {
		t.Fatalf("expected 0 errors but got: %s", err)
	}
	expected := map[string]interface{}{
		"simple": "a_value_v1",
		"second": "a_second_value_v2",
	}
	if !reflect.DeepEqual(res, expected) {
		t.Errorf("expected: %s, got: %s.", expected, res)
	}
}

func TestAzGetSecretsWithError(t *testing.T) {
	var keyVault = newAzureKeyVaultBackendMock("fetch_error")
	var _, err = keyVault.GetSecrets("keyvault", "", nil)
	if err == nil {
		t.Fatalf("expected 1 errors but got nil")
	}
	expected := errors.New("fetch error")
	if !reflect.DeepEqual(err, expected) {
		t.Errorf("expected err: %s, got: %s.", expected, err)
	}
}

func TestAzGetSecretsWithErrorOnGetSecret(t *testing.T) {
	var keyVault = newAzureKeyVaultBackendMock("get_secret_error")
	var _, err = keyVault.GetSecrets("keyvault", "", nil)
	if err == nil {
		t.Fatalf("expected 1 errors but got nil")
	}
	expected := errors.New("secret not found")
	if !reflect.DeepEqual(err, expected) {
		t.Errorf("expected err: %s, got: %s.", expected, err)
	}
}

func TestAzGetSecretsBuilderError(t *testing.T) {
	var keyVault = &backends.AzureKeyVault{
		Credential: nil,
		ClientBuilder: func(vaultURL string, credential azcore.TokenCredential, options *azsecrets.ClientOptions) (backends.AzSecretsClient, error) {
			return nil, errors.New("boom")
		},
	}
	var _, err = keyVault.GetSecrets("keyvault", "", nil)
	if err == nil {
		t.Fatalf("expected 1 errors but got nil")
	}
	expected := errors.New("boom")
	if !reflect.DeepEqual(err, expected) {
		t.Errorf("expected err: %s, got: %s.", expected, err)
	}
}

func TestAzGetSecretsVersionV1(t *testing.T) {
	var keyVault = newAzureKeyVaultBackendMock("")
	var res, err = keyVault.GetSecrets("keyvault", "v1", nil)

	if err != nil {
		t.Fatalf("expected 0 errors but got: %s", err)
	}
	expected := map[string]interface{}{
		"simple": "a_value_v1",
	}
	if !reflect.DeepEqual(res, expected) {
		t.Errorf("expected: %s, got: %s.", expected, res)
	}
}

func TestAzGetSecretsVersionV2(t *testing.T) {
	var keyVault = newAzureKeyVaultBackendMock("")
	var res, err = keyVault.GetSecrets("keyvault", "v2", nil)

	if err != nil {
		t.Fatalf("expected 0 errors but got: %s", err)
	}
	expected := map[string]interface{}{
		"simple": "a_value_v2",
		"second": "a_second_value_v2",
	}
	if !reflect.DeepEqual(res, expected) {
		t.Errorf("expected: %s, got: %s.", expected, res)
	}
}
