package vault

import (
	"fmt"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"github.com/hashicorp/vault/api"
)

const (
	githubMountPath = "auth/github"
)

// GithubAuth is a struct for working with Vault that uses the Github Auth method
type GithubAuth struct {
	AccessToken string
	MountPath   string
}

// NewGithubAuth initializes a new GithubAuth with token
func NewGithubAuth(token, mountPath string) *GithubAuth {
	githubAuth := &GithubAuth{
		AccessToken: token,
		MountPath:   githubMountPath,
	}
	if mountPath != "" {
		githubAuth.MountPath = mountPath
	}

	return githubAuth
}

// Authenticate authenticates with Vault and returns a token
func (g *GithubAuth) Authenticate(vaultClient *api.Client) error {
	payload := map[string]interface{}{
		"token": g.AccessToken,
	}

	utils.VerboseToStdErr("Hashicorp Vault authenticating with Github token %s", g.AccessToken)
	data, err := vaultClient.Logical().Write(fmt.Sprintf("%s/login", g.MountPath), payload)
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
