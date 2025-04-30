package backends

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/example/helpers"
	ocism "github.com/oracle/oci-go-sdk/v65/secrets"
	ocivault "github.com/oracle/oci-go-sdk/v65/vault"
)

var OCIPath, _ = regexp.Compile(`^ocivault(?:/(.+))?$`)
var OCISecretVersion, _ = regexp.Compile(`^\d+$`)

type OCISecretIface interface {
	GetSecretBundleByName(ctx context.Context, request ocism.GetSecretBundleByNameRequest) (response ocism.GetSecretBundleByNameResponse, err error)
}

type OCIVaultIface interface {
	ListSecrets(ctx context.Context, request ocivault.ListSecretsRequest) (response ocivault.ListSecretsResponse, err error)
	ListSecretVersions(ctx context.Context, request ocivault.ListSecretVersionsRequest) (response ocivault.ListSecretVersionsResponse, err error)
}

// OCIVault is a struct for working with a OCI Vault backend
type OCIVault struct {
	secretClient OCISecretIface
	vaultClient OCIVaultIface
	vaultId string
	compartmentId string
}

// NewOCIVaultBackend initializes a new OCI Vault backend
func NewOCIVaultBackend(secret_client OCISecretIface,vault_client OCIVaultIface, vault_id string, compartment_id string ) *OCIVault {
	return &OCIVault{
		secretClient: secret_client,
		vaultClient: vault_client,
		vaultId: vault_id,
		compartmentId: compartment_id,
	}
}

// Login does nothing as a "login" is handled on the instantiation of the oci sdk
func (oci *OCIVault) Login() error {
	return nil
}


// Iterate Secret Version
func (oci *OCIVault) CheckSecretVersion(secret_version_req ocivault.ListSecretVersionsRequest, version int64 ) (bool,error) {

	version_found := false

	listSecretVersionFunc := func(request ocivault.ListSecretVersionsRequest) (ocivault.ListSecretVersionsResponse, error) {
		return oci.vaultClient.ListSecretVersions(context.Background(), request)
	}

	for r, err := listSecretVersionFunc(secret_version_req); ; r, err = listSecretVersionFunc(secret_version_req) {
		if err != nil{
			return version_found,err
		}

		for _, secret_ver_summary := range r.Items {
			if secret_ver_summary.VersionNumber != nil && *secret_ver_summary.VersionNumber == version {
					version_found=true
					break
				}
		}

		if r.OpcNextPage != nil && !version_found {
			// if there are more items in next page, fetch items from next page
			secret_version_req.Page = r.OpcNextPage
		} else {
			// no more result, break the loop
			break
		}
	}

	return version_found,nil
}

// GetSecrets gets secrets from OCI Vault and returns the formatted data
// For OCI  Vault, the path is of format `ocivault/secretname`

func (oci *OCIVault) GetSecrets(kvpath string, version string, annotations map[string]string) (map[string]interface{}, error) {

	var secretName *string = nil
	matches := OCIPath.FindStringSubmatch(kvpath)
	if len(matches) == 0 {
		return nil, fmt.Errorf("path is not in the correct format (ocivault/) for OCI vault: %s", kvpath)
	}

	if len(matches) > 1 && matches[1] != "" {
		secretName = &matches[1] // Capture secret name if it exists
	} 

	list_secrets_req := ocivault.ListSecretsRequest{
		CompartmentId: common.String(oci.compartmentId),
		SortBy:         ocivault.ListSecretsSortByName,
		VaultId:        common.String(oci.vaultId),
		Name: 			secretName,
		Limit:          common.Int(100),
		LifecycleState: ocivault.SecretSummaryLifecycleStateActive,
		SortOrder:      ocivault.ListSecretsSortOrderDesc}

	listSecretsFunc := func(request ocivault.ListSecretsRequest) (ocivault.ListSecretsResponse, error) {
		return oci.vaultClient.ListSecrets(context.Background(), request)
	}

	data := make(map[string]interface{})

	for r, err := listSecretsFunc(list_secrets_req); ; r, err = listSecretsFunc(list_secrets_req) {
		helpers.FatalIfError(err)

		for _, secret := range r.Items {
			req := ocism.GetSecretBundleByNameRequest{
				VaultId:       common.String(oci.vaultId),
				SecretName:    common.String(*secret.SecretName)}

			if version != "" && !strings.EqualFold(version, "latest") {
				isPositiveInteger := OCISecretVersion.MatchString(version)
				if !isPositiveInteger {
					return nil, fmt.Errorf("version string must contain only positive integers")
				}
				secret_version, err := strconv.ParseInt(version, 10, 64)
				helpers.FatalIfError(err)

				list_secret_version_req := ocivault.ListSecretVersionsRequest{
					SecretId:     common.String(*secret.Id),
					SortBy:       ocivault.ListSecretVersionsSortByVersionNumber,
					SortOrder:    ocivault.ListSecretVersionsSortOrderAsc,
					Limit:        common.Int(100)}
				
				version_found,err := oci.CheckSecretVersion(list_secret_version_req,secret_version)
				helpers.FatalIfError(err)
				if version_found {
					req.VersionNumber = common.Int64(secret_version)
				} else{
					req.Stage = ocism.GetSecretBundleByNameStageLatest
				}
			} else {
				req.Stage = ocism.GetSecretBundleByNameStageLatest
			}

			resp, err := oci.secretClient.GetSecretBundleByName(context.Background(), req)
			helpers.FatalIfError(err)

			secretContent := resp.SecretBundle.SecretBundleContent.(ocism.Base64SecretBundleContentDetails)
			encodedSecret := *secretContent.Content

			data[*secret.SecretName] = string(encodedSecret)
	}

		if r.OpcNextPage != nil {
			// if there are more items in next page, fetch items from next page
			list_secrets_req.Page = r.OpcNextPage
		} else {
			// no more result, break the loop
			break
		}
	}

	return data, nil
}


// GetIndividualSecret will get the specific secret (placeholder) from the SM backend
// For OCI Vault, the path is of format `ocivault/secretname`
// So, we use GetSecrets and extract the specific placeholder we want
func (oci *OCIVault) GetIndividualSecret(kvpath, secret, version string, annotations map[string]string) (interface{}, error) {
	secretpath := kvpath + "/" + secret
	data, err := oci.GetSecrets(secretpath, version, annotations)
	if err != nil {
		return nil, err
	}
	return data[secret], nil
}
