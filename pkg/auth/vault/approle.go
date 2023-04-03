package vault

import (
	"fmt"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"github.com/hashicorp/vault/api"
)

const (
	approleMountPath = "auth/approle"
)

// AppRoleAuth is a struct for working with Vault that uses AppRole
type AppRoleAuth struct {
	RoleID    string
	SecretID  string
	MountPath string
}

// NewAppRoleAuth initalizes a new AppRolAuth with role id and secret id
func NewAppRoleAuth(roleID, secretID, mountPath string) *AppRoleAuth {
	appRoleAuth := &AppRoleAuth{
		RoleID:    roleID,
		SecretID:  secretID,
		MountPath: approleMountPath,
	}
	if mountPath != "" {
		appRoleAuth.MountPath = mountPath
	}

	return appRoleAuth
}

// Authenticate authenticates with Vault using App Role and returns a token
func (a *AppRoleAuth) Authenticate(vaultClient *api.Client) error {
	payload := map[string]interface{}{
		"role_id":   a.RoleID,
		"secret_id": a.SecretID,
	}

	utils.VerboseToStdErr("Hashicorp Vault authenticating with role ID %s and secret ID %s at path %s", a.RoleID, a.SecretID, a.MountPath)
	data, err := vaultClient.Logical().Write(fmt.Sprintf("%s/login", a.MountPath), payload)
	if err != nil {
		return err
	}

	utils.VerboseToStdErr("Hashicorp Vault authentication response: %v", data)

	// If we cannot write the Vault token, we'll just have to login next time. Nothing showstopping.
	err = utils.SetToken(vaultClient, data.Auth.ClientToken)
	if err != nil {
		utils.VerboseToStdErr("Hashicorp Vault cannot cache token for future runs: %v", err)
	}

	return nil
}
