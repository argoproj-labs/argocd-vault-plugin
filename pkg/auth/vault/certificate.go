package vault

import (
	"fmt"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"github.com/hashicorp/vault/api"
	"os"
)

const (
	certificateMountPath = "auth/cert"
)

// CertificateAuth is a struct for working with Vault that uses certificate authentication
type CertificateAuth struct {
	Certificate string
	Key         string
	MountPath   string
}

// NewCertificateAuth initalizes a new CertificateAuth with cert & key
func NewCertificateAuth(cert, key, mountPath string) *CertificateAuth {
	certificateAuth := &CertificateAuth{
		Certificate: cert,
		Key:         key,
		MountPath:   certificateMountPath,
	}
	if mountPath != "" {
		certificateAuth.MountPath = mountPath
	}

	return certificateAuth
}

// Authenticate authenticates with Vault using userpass and returns a token
func (a *CertificateAuth) Authenticate(vaultClient *api.Client) error {
	err := utils.LoginWithCachedToken(vaultClient)
	if err != nil {
		utils.VerboseToStdErr("Hashicorp Vault cannot retrieve cached token: %v. Generating a new one", err)
	} else {
		return nil
	}

	payload := map[string]interface{}{}

	tempCrt, err := os.CreateTemp("", "vault_cert")
	if err != nil {
		return err
	}
	if _, err := tempCrt.WriteString(a.Certificate); err != nil {
		return err
	}
	defer os.Remove(tempCrt.Name())

	tempKey, err := os.CreateTemp("", "vault_key")
	if err != nil {
		return err
	}

	if _, err := tempKey.WriteString(a.Key); err != nil {
		return err
	}
	defer os.Remove(tempKey.Name())

	// Clone Client with new TLS Settings
	apiClientConfig := vaultClient.CloneConfig()

	tlsConfig := &api.TLSConfig{
		ClientKey:  tempKey.Name(),
		ClientCert: tempCrt.Name(),
	}

	err = apiClientConfig.ConfigureTLS(tlsConfig)
	if err != nil {
		return err
	}

	certVaultClient, err := api.NewClient(apiClientConfig)

	if err != nil {
		return err
	}

	utils.VerboseToStdErr("Hashicorp Vault authenticating with certificate")

	certVaultClient.ClearToken()
	data, err := certVaultClient.Logical().Write(fmt.Sprintf("%s/login", a.MountPath), payload)
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
