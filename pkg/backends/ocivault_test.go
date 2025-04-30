package backends_test

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/backends"
	"github.com/oracle/oci-go-sdk/v65/common"
	ocism "github.com/oracle/oci-go-sdk/v65/secrets"
	ocivault "github.com/oracle/oci-go-sdk/v65/vault"
)

type mockOCISecretClient struct {
	backends.OCISecretIface
}

type mockOCIVaultClient struct {
	backends.OCIVaultIface
}

func (m *mockOCISecretClient) GetSecretBundleByName(ctx context.Context, request ocism.GetSecretBundleByNameRequest) (response ocism.GetSecretBundleByNameResponse, err error) {
	data := ocism.GetSecretBundleByNameResponse{}

	switch *request.VaultId {
	case "test_vaultid":
		if request.VersionNumber == nil || request.Stage == ocism.GetSecretBundleByNameStageLatest {
			content := "current-value"
			secretBundle := ocism.Base64SecretBundleContentDetails{
				Content: &content,
			}
			data.SecretBundle.SecretBundleContent = secretBundle
		} else {
			content := "previous-value"
			secretBundle := ocism.Base64SecretBundleContentDetails{
				Content: &content,
			}
			data.SecretBundle.SecretBundleContent = secretBundle
		}

	}

	return data, nil
}

func (m *mockOCIVaultClient) ListSecrets(ctx context.Context, request ocivault.ListSecretsRequest) (response ocivault.ListSecretsResponse, err error) {
	data := ocivault.ListSecretsResponse{}
	secretname := "secretname"
	secretid := "secretid"
	secretsummary := ocivault.SecretSummary{
		SecretName: &secretname,
		Id: &secretid,

	}
	data.Items = append(data.Items, secretsummary)
	return data, nil
}

func (m *mockOCIVaultClient) ListSecretVersions(ctx context.Context, request ocivault.ListSecretVersionsRequest) (response ocivault.ListSecretVersionsResponse, err error) {
	data := ocivault.ListSecretVersionsResponse{}
	version_number := int64(1)
	secret_ver_summary := ocivault.SecretVersionSummary{
		VersionNumber: &version_number,

	}
	data.Items = append(data.Items, secret_ver_summary)

	return data, nil
}

func TestOCIVaultGetSecrets(t *testing.T) {
	sm := backends.NewOCIVaultBackend(&mockOCISecretClient{},&mockOCIVaultClient{},"test_vaultid","test_compartmentid")

	t.Run("Get latest secrets", func(t *testing.T) {
		data, err := sm.GetSecrets("ocivault/secretname", "latest", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"secretname": "current-value",
		}

		if !reflect.DeepEqual(expected, data) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}
	})

	t.Run("Get secrets without version", func(t *testing.T) {
		data, err := sm.GetSecrets("ocivault/secretname", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"secretname": "current-value",
		}

		if !reflect.DeepEqual(expected, data) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}
	})

	t.Run("Get secrets with existing version", func(t *testing.T) {
		data, err := sm.GetSecrets("ocivault/secretname", "1", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"secretname": "previous-value",
		}

		if !reflect.DeepEqual(expected, data) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}
	})

	t.Run("Get secrets with non existing version", func(t *testing.T) {
		data, err := sm.GetSecrets("ocivault/secretname", "5", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"secretname": "current-value",
		}

		if !reflect.DeepEqual(expected, data) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}
	})

	t.Run("OCI latest GetIndividualSecret", func(t *testing.T) {
		secret, err := sm.GetIndividualSecret("ocivault/secretname", "secretname", "latest", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := "current-value"
		if !reflect.DeepEqual(expected, secret) {
			t.Errorf("expected: %s, got: %s.", expected, secret)
		}
	})

	t.Run("OCI GetIndividualSecret without version", func(t *testing.T) {
		secret, err := sm.GetIndividualSecret("ocivault/secretname", "secretname", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := "current-value"
		if !reflect.DeepEqual(expected, secret) {
			t.Errorf("expected: %s, got: %s.", expected, secret)
		}
	})

	t.Run("OCI GetIndividualSecret with existing version", func(t *testing.T) {
		secret, err := sm.GetIndividualSecret("ocivault/secretname", "secretname", "1", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := "previous-value"
		if !reflect.DeepEqual(expected, secret) {
			t.Errorf("expected: %s, got: %s.", expected, secret)
		}
	})

	t.Run("OCI GetIndividualSecret with non existing version", func(t *testing.T) {
		secret, err := sm.GetIndividualSecret("ocivault/secretname", "secretname", "5", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := "current-value"
		if !reflect.DeepEqual(expected, secret) {
			t.Errorf("expected: %s, got: %s.", expected, secret)
		}
	})
	t.Run("OCI Invalid Path", func(t *testing.T) {
		_, err := sm.GetSecrets("ibmcloud/arbitrary/secrets/groups/some-group", "", map[string]string{})
		if err == nil {
			t.Fatalf("expected error")
		}

		expectedErr := "path is not in the correct format"
		if !strings.Contains(err.Error(), expectedErr) {
			t.Fatalf("Expected error to have %s but said %s", expectedErr, err)
		}
	})

	t.Run("OCI Invalid Version Number", func(t *testing.T) {
		_, err := sm.GetSecrets("ocivault/secretname", "eleven", map[string]string{})
		if err == nil {
			t.Fatalf("expected error")
		}

		expectedErr := "version string must contain only positive integers"
		if !strings.Contains(err.Error(), expectedErr) {
			t.Fatalf("Expected error to have %s but said %s", expectedErr, err)
		}
	})

	t.Run("OCI Negative Version Number", func(t *testing.T) {
		_, err := sm.GetSecrets("ocivault/secretname", "-1", map[string]string{})
		if err == nil {
			t.Fatalf("expected error")
		}

		expectedErr := "version string must contain only positive integers"
		if !strings.Contains(err.Error(), expectedErr) {
			t.Fatalf("Expected error to have %s but said %s", expectedErr, err)
		}
	})

	t.Run("OCI Float Version Number", func(t *testing.T) {
		_, err := sm.GetSecrets("ocivault/secretname", "1.0", map[string]string{})
		if err == nil {
			t.Fatalf("expected error")
		}

		expectedErr := "version string must contain only positive integers"
		if !strings.Contains(err.Error(), expectedErr) {
			t.Fatalf("Expected error to have %s but said %s", expectedErr, err)
		}
	})

	t.Run("OCI Check Version Number", func(t *testing.T) {
		list_secret_version_req := ocivault.ListSecretVersionsRequest{
			SecretId:     common.String("secretname"),
			SortBy:       ocivault.ListSecretVersionsSortByVersionNumber,
			SortOrder:    ocivault.ListSecretVersionsSortOrderAsc,
			Limit:        common.Int(100)}		
		data, err := sm.CheckSecretVersion(list_secret_version_req,int64(1))
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := true

		if !reflect.DeepEqual(expected, data) {
			t.Errorf("expected: %v, got: %v.", expected, data)
		}
	})

}
