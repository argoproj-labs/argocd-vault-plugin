package vault

import (
	"fmt"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"github.com/hashicorp/vault/api"
)

const (
	userpassMountPath = "auth/userpass"
)

// UserPassAuth is a struct for working with Vault that uses Username & Password
type UserPassAuth struct {
	Username  string
	Password  string
	MountPath string
}

// NewUserPassAuth initalizes a new NewUserPassAuth with username & password
func NewUserPassAuth(username, password, mountPath string) *UserPassAuth {
	userpassAuth := &UserPassAuth{
		Username:  username,
		Password:  password,
		MountPath: userpassMountPath,
	}
	if mountPath != "" {
		userpassAuth.MountPath = mountPath
	}

	return userpassAuth
}

// Authenticate authenticates with Vault using userpass and returns a token
func (a *UserPassAuth) Authenticate(vaultClient *api.Client) error {
	payload := map[string]interface{}{
		"password": a.Password,
	}

	utils.VerboseToStdErr("Hashicorp Vault authenticating with username %s and password %s", a.Username, a.Password)
	data, err := vaultClient.Logical().Write(fmt.Sprintf("%s/login/%s", a.MountPath, a.Username), payload)
	if err != nil {
		return err
	}

	utils.VerboseToStdErr("Hashicorp Vault authentication response: %v", data)

	// If we cannot write the Vault token, we'll just have to login next time. Nothing showstopping.
	if err = utils.SetToken(vaultClient, data.Auth.ClientToken); err != nil {
		utils.VerboseToStdErr("Hashicorp Vault cannot cache token for future runs: %v", err)
	}

	return nil
}
