package vault

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/IBM/argocd-vault-plugin/pkg/utils"
	"github.com/hashicorp/vault/api"
)

// K8sAuth TODO
type K8sAuth struct {
	MountPath string
	TokenPath string
	Role      string
}

// NewK8sAuth TODO
func NewK8sAuth(role, mountPath, tokenPath string) *K8sAuth {
	k8sAuth := &K8sAuth{
		Role:      role,
		MountPath: mountPath,
		TokenPath: tokenPath,
	}

	return k8sAuth
}

// Authenticate authenticates with Vault via K8s and returns a token
func (k *K8sAuth) Authenticate(vaultClient *api.Client) error {
	token, err := k.getJWT()
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"role": k.Role,
		"jwt":  token,
	}

	data, err := vaultClient.Logical().Write(fmt.Sprintf("%s/login", k.MountPath), payload)
	if err != nil {
		return err
	}

	// If we cannot write the Vault token, we'll just have to login next time. Nothing showstopping.
	err = utils.SetToken(vaultClient, data.Auth.ClientToken)
	if err != nil {
		print(err)
	}

	return nil
}

func (k *K8sAuth) getJWT() (string, error) {
	tokenFilePath := "/var/run/secrets/kubernetes.io/serviceaccount/token"
	if k.TokenPath != "" {
		tokenFilePath = k.TokenPath
	}

	f, err := os.Open(tokenFilePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	contentBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(contentBytes)), nil
}
