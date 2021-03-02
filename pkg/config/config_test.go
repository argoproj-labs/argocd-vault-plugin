package config_test

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/IBM/argocd-vault-plugin/pkg/config"
	"github.com/IBM/argocd-vault-plugin/pkg/utils"
	"github.com/spf13/viper"
)

func TestNewConfig(t *testing.T) {
	testCases := []struct {
		environment  map[string]interface{}
		expectedType string
	}{
		{
			map[string]interface{}{
				"AVP_TYPE":         "vault",
				"AVP_AUTH_TYPE":    "github",
				"AVP_GITHUB_TOKEN": "token",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":      "vault",
				"AVP_AUTH_TYPE": "approle",
				"AVP_ROLE_ID":   "role_id",
				"AVP_SECRET_ID": "secret_id",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":            "vault",
				"AVP_AUTH_TYPE":       "k8s",
				"AVP_K8S_MOUNT_POINT": "mount_point",
				"AVP_K8S_ROLE":        "role",
				"AVP_K8S_TOKEN_PATH":  "toke_path",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":      "vault",
				"AVP_AUTH_TYPE": "k8s",
				"AVP_K8S_ROLE":  "role",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":            "vault",
				"AVP_AUTH_TYPE":       "k8s",
				"AVP_K8S_MOUNT_POINT": "mount_point",
				"AVP_K8S_ROLE":        "role",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":        "secretmanager",
				"AVP_AUTH_TYPE":   "iam",
				"AVP_IBM_API_KEY": "token",
			},
			"*backends.IBMSecretManager",
		},
	}
	for _, tc := range testCases {
		for k, v := range tc.environment {
			os.Setenv(k, v.(string))
		}
		viper := viper.New()
		config, err := config.New(viper, utils.DefaultHttpClient())
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		xType := fmt.Sprintf("%T", config.Backend)
		if xType != tc.expectedType {
			t.Errorf("expected: %s, got: %s.", tc.expectedType, xType)
		}
		for k := range tc.environment {
			os.Unsetenv(k)
		}
	}
}

func TestVaultNamespace(t *testing.T) {
	env := map[string]interface{}{
		"AVP_TYPE":            "vault",
		"AVP_AUTH_TYPE":       "github",
		"AVP_GITHUB_TOKEN":    "token",
		"AVP_VAULT_NAMESPACE": "ns1",
	}

	for k, v := range env {
		os.Setenv(k, v.(string))
	}

	viper := viper.New()
	cf, err := config.New(viper, utils.DefaultHttpClient())
	if err != nil {
		t.Fatalf("expected 0 errors but got: %s", err)
	}

	headers := cf.VaultClient.Headers()
	if headers.Get("X-Vault-Namespace") != "ns1" {
		t.Errorf("expected X-Vault-Namespace to be %s, got %s", "ns1", headers.Get("X-Vault-Namespace"))
	}

	for k := range env {
		os.Unsetenv(k)
	}
}

func TestVaultTLS(t *testing.T) {
	fakeCert := `-----BEGIN CERTIFICATE-----
MIICsjCCAZoCCQCpRK9HNTzoXzANBgkqhkiG9w0BAQsFADAbMQswCQYDVQQGEwJV
UzEMMAoGA1UECgwDSUJNMB4XDTIxMDMwMTE1NDU1M1oXDTIxMDMzMTE1NDU1M1ow
GzELMAkGA1UEBhMCVVMxDDAKBgNVBAoMA0lCTTCCASIwDQYJKoZIhvcNAQEBBQAD
ggEPADCCAQoCggEBANawTlNFjatQCP/ydFgUnEaF01/67z514i5v/LDEPgEvupkd
Z/GPueZXvIu65RS3DcZKTBeg5ACIwp6X9zJenCy3NWXHq5ro0hfNNG0F4GgjUMAH
V4wz3Oi+LsrnybPLcD3U8PXhRytsAQNqUG3Cx2gd0i+knaIy/WHgUlBJiNJLoOlW
/wfNGNDuZcOWs68kdf2mrBEPMOWGRpC2lBw4BeUEvPqZAs3eNVRWETL3TkkZVDaP
CkIwY44xOSdZnx0c6JOTxWu11caX33sCGWZVwdIlPlWSdHk3ktXjWHIcizCT6GpX
wKEkhXI6hPJpWTa7//RFJwvZm28F732Zbcb1kEMCAwEAATANBgkqhkiG9w0BAQsF
AAOCAQEAvi9GSPRkkMhC0B2L9HWYSuaEk7VIGvrRmRNL/IZXB8KRBV1kF913Mdy5
bEwEWF6AKL2lvbSRW9QIGoBn777ZHj+vCxdWbic7uLNIuvMFX6CvUM7uCj4+9tjy
+oaiirgSgu6K8aq8b/nPwN2b924YWadhxsTlu/vRDBnqtmNc82zsM32wF1GA38Yx
intZBFinXKrfHBqwJlWRxTRQtnx1UjotE0Hxo4rxaSSWxPOFMk44id0iZ9+QVhAu
4LJiL6ZSRnhaBhiTdtbTMPfFjzAlpMlu6RsA9X8OBbm6IklE0kSw4lVQ0fgzSjR7
1ul5Nue2ISTGUBUvkLR59+DeHqCNBQ==
-----END CERTIFICATE-----`

	err := ioutil.WriteFile("/tmp/test-cert.crt", []byte(fakeCert), 0644)
	if err != nil {
		t.Fatal(err)
	}

	env := map[string]interface{}{
		"AVP_TYPE":              "vault",
		"AVP_AUTH_TYPE":         "github",
		"AVP_GITHUB_TOKEN":      "token",
		"AVP_VAULT_CACERT":      "/tmp/test-cert.crt",
		"AVP_VAULT_CAPATH":      "/tmp/test-cert.crt",
		"AVP_VAULT_SKIP_VERIFY": "true",
	}

	for k, v := range env {
		os.Setenv(k, v.(string))
	}

	var tlsClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	var transport http.RoundTripper = &http.Transport{
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     tlsClientConfig,
	}

	var httpClient = &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	viper := viper.New()
	_, err = config.New(viper, httpClient)
	if err != nil {
		t.Fatalf("expected 0 errors but got: %s", err)
	}

	if len(tlsClientConfig.RootCAs.Subjects()) != 1 {
		t.Errorf("expected cert count to be 1, got %v", len(tlsClientConfig.RootCAs.Subjects()))
	}

	if tlsClientConfig.InsecureSkipVerify != true {
		t.Errorf("expected insecure skip verify to be true, got %v", tlsClientConfig.InsecureSkipVerify)
	}

	for k := range env {
		os.Unsetenv(k)
	}
}

func TestNewConfigNoType(t *testing.T) {
	viper := viper.New()
	_, err := config.New(viper, utils.DefaultHttpClient())
	expectedError := "Must provide a supported Vault Type"

	if err.Error() != expectedError {
		t.Errorf("expected error %s to be thrown, got %s", expectedError, err)
	}
}

func TestNewConfigNoAuthType(t *testing.T) {
	os.Setenv("AVP_TYPE", "vault")
	viper := viper.New()
	_, err := config.New(viper, utils.DefaultHttpClient())
	expectedError := "Must provide a supported Authentication Type"

	if err.Error() != expectedError {
		t.Errorf("expected error %s to be thrown, got %s", expectedError, err)
	}
	os.Unsetenv("AVP_TYPE")
}

func TestNewConfigMissingParameter(t *testing.T) {
	testCases := []struct {
		environment  map[string]interface{}
		expectedType string
	}{
		{
			map[string]interface{}{
				"AVP_TYPE":      "vault",
				"AVP_AUTH_TYPE": "github",
				"AVP_GH_TOKEN":  "token",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":      "vault",
				"AVP_AUTH_TYPE": "approle",
				"AVP_ROLEID":    "role_id",
				"AVP_SECRET_ID": "secret_id",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":      "vault",
				"AVP_AUTH_TYPE": "k8s",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":        "secretmanager",
				"AVP_AUTH_TYPE":   "iam",
				"AVP_IAM_API_KEY": "token",
			},
			"*backends.IBMSecretManager",
		},
	}
	for _, tc := range testCases {
		for k, v := range tc.environment {
			os.Setenv(k, v.(string))
		}
		viper := viper.New()
		_, err := config.New(viper, utils.DefaultHttpClient())
		if err == nil {
			t.Fatalf("%s should not instantiate", tc.expectedType)
		}
		for k := range tc.environment {
			os.Unsetenv(k)
		}
	}

}
