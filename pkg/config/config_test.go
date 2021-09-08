package config_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/IBM/argocd-vault-plugin/pkg/config"
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
				"AVP_AUTH_TYPE": "token",
				"VAULT_TOKEN":   "token",
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
				"AVP_TYPE":             "ibmsecretsmanager",
				"AVP_IBM_API_KEY":      "token",
				"AVP_IBM_INSTANCE_URL": "http://ibm",
			},
			"*backends.IBMSecretsManager",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":        "ibmsecretsmanager",
				"AVP_IBM_API_KEY": "token",
				"VAULT_ADDR":      "http://ibm",
			},
			"*backends.IBMSecretsManager",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":              "awssecretsmanager",
				"AWS_REGION":            "us-west-1",
				"AWS_ACCESS_KEY_ID":     "id",
				"AWS_SECRET_ACCESS_KEY": "key",
			},
			"*backends.AWSSecretsManager",
		},
		{ // auth via web identity federation is also possible
			map[string]interface{}{
				"AVP_TYPE":                    "awssecretsmanager",
				"AWS_REGION":                  "us-west-1",
				"AWS_WEB_IDENTITY_TOKEN_FILE": "/var/run/secrets/eks.amazonaws.com/serviceaccount/token",
				"AWS_ROLE_ARN":                "arn:aws:iam::111111111:role/argocd-repo-server-secretsmanager-my-cluster",
			},
			"*backends.AWSSecretsManager",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":                       "gcpsecretmanager",
				"GOOGLE_APPLICATION_CREDENTIALS": "../../fixtures/input/gac.json",
			},
			"*backends.GCPSecretManager",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":            "azurekeyvault",
				"AZURE_TENANT_ID":     "test",
				"AZURE_CLIENT_ID":     "test",
				"AZURE_CLIENT_SECRET": "test",
			},
			"*backends.AzureKeyVault",
		},
	}
	for _, tc := range testCases {
		for k, v := range tc.environment {
			os.Setenv(k, v.(string))
		}
		viper := viper.New()
		config, err := config.New(viper, &config.Options{})
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

func TestNewConfigNoType(t *testing.T) {
	viper := viper.New()
	_, err := config.New(viper, &config.Options{})
	expectedError := "Must provide a supported Vault Type"

	if err.Error() != expectedError {
		t.Errorf("expected error %s to be thrown, got %s", expectedError, err)
	}
}

func TestNewConfigNoAuthType(t *testing.T) {
	os.Setenv("AVP_TYPE", "vault")
	viper := viper.New()
	_, err := config.New(viper, &config.Options{})
	expectedError := "Must provide a supported Authentication Type"

	if err.Error() != expectedError {
		t.Errorf("expected error %s to be thrown, got %s", expectedError, err)
	}
	os.Unsetenv("AVP_TYPE")
}

// Helper function that captures log output from a function call into a string
// Adapted from https://stackoverflow.com/a/26806093/170154
func captureOutput(f func()) string {
	var buf bytes.Buffer
	flags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0) // don't include any date or time in the logging messages
	f()
	log.SetOutput(os.Stderr)
	log.SetFlags(flags)
	return buf.String()
}

func TestNewConfigAwsRegionWarning(t *testing.T) {
	testCases := []struct {
		environment  map[string]interface{}
		expectedType string
		expectedLog  string
	}{
		{ // this test issues a warning for missing AWS_REGION env var
			map[string]interface{}{
				"AVP_TYPE":              "awssecretsmanager",
				"AWS_ACCESS_KEY_ID":     "id",
				"AWS_SECRET_ACCESS_KEY": "key",
			},
			"*backends.AWSSecretsManager",
			"Warning: AWS_REGION env var not set, using AWS region us-east-2.\n",
		},
		{ // no warning is issued
			map[string]interface{}{
				"AVP_TYPE":              "awssecretsmanager",
				"AWS_REGION":            "us-west-1",
				"AWS_ACCESS_KEY_ID":     "id",
				"AWS_SECRET_ACCESS_KEY": "key",
			},
			"*backends.AWSSecretsManager",
			"",
		},
	}

	for _, tc := range testCases {
		for k, v := range tc.environment {
			os.Setenv(k, v.(string))
		}
		viper := viper.New()

		output := captureOutput(func() {
			config, err := config.New(viper, &config.Options{})
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			xType := fmt.Sprintf("%T", config.Backend)
			if xType != tc.expectedType {
				t.Errorf("expected: %s, got: %s.", tc.expectedType, xType)
			}
		})

		if output != tc.expectedLog {
			t.Errorf("Unexpected warning issued. Expected: %s, actual: %s", tc.expectedLog, output)
		}

		for k := range tc.environment {
			os.Unsetenv(k)
		}
	}
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
				"AVP_AUTH_TYPE": "token",
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
				"AVP_TYPE":        "ibmsecretsmanager",
				"AVP_IAM_API_KEY": "token",
			},
			"*backends.IBMSecretsManager",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":   "ibmsecretsmanager",
				"VAULT_ADDR": "http://vault",
			},
			"*backends.IBMSecretsManager",
		},
		{ //  WebIdentityEmptyRoleARNErr will occur if 'AWS_WEB_IDENTITY_TOKEN_FILE' was set but 'AWS_ROLE_ARN' was not set.
			map[string]interface{}{
				"AVP_TYPE":                    "awssecretsmanager",
				"AWS_REGION":                  "us-west-1",
				"AWS_WEB_IDENTITY_TOKEN_FILE": "/var/run/secrets/eks.amazonaws.com/serviceaccount/token",
			},
			"*backends.AWSSecretsManager",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":            "azurekeyvault",
				"AZURE_TENANT_ID":     "test",
				"AZURE_CLIENT_ID":     "test",
			},
			"*backends.AzureKeyVault",
		},
	}
	for _, tc := range testCases {
		for k, v := range tc.environment {
			os.Setenv(k, v.(string))
		}
		viper := viper.New()
		_, err := config.New(viper, &config.Options{})
		if err == nil {
			t.Fatalf("%s should not instantiate", tc.expectedType)
		}
		for k := range tc.environment {
			os.Unsetenv(k)
		}
	}

}

func TestExternalConfig(t *testing.T) {
	os.Setenv("AVP_TYPE", "vault")
	viper := viper.New()
	viper.SetDefault("VAULT_ADDR", "http://my-vault:8200/")
	config.New(viper, &config.Options{})
	if os.Getenv("VAULT_ADDR") != "http://my-vault:8200/" {
		t.Errorf("expected VAULT_ADDR env to be set from external config, was instead: %s", os.Getenv("VAULT_ADDR"))
	}
	os.Unsetenv("AVP_TYPE")
	os.Unsetenv("VAULT_ADDR")
}

const avpConfig = `AVP_TYPE: awssecretsmanager
AWS_ACCESS_KEY_ID: AKIAIOSFODNN7EXAMPLE
AWS_SECRET_ACCESS_KEY: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
AWS_REGION: us-west-2`

var expectedEnvVars = map[string]string{
	"AVP_TYPE":              "", // shouldn't be an env var
	"AWS_ACCESS_KEY_ID":     "AKIAIOSFODNN7EXAMPLE",
	"AWS_SECRET_ACCESS_KEY": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	"AWS_REGION":            "us-west-2",
}

func TestExternalConfigAWS(t *testing.T) {
	// Test setting AWS_* env variables from external AVP config, note setting
	// env vars is necessary to pass AVP config entries to the AWS golang SDK
	tmpFile, err := ioutil.TempFile("", "avpConfig.*.yaml")
	if err != nil {
		t.Errorf("Cannot create temporary file %s", err)
	}

	defer os.Remove(tmpFile.Name()) // clean up the file afterwards

	if _, err = tmpFile.WriteString(avpConfig); err != nil {
		t.Errorf("Failed to write to temporary file %s", err)
	}

	viper := viper.New()
	if _, err = config.New(viper, &config.Options{ConfigPath: tmpFile.Name()}); err != nil {
		t.Errorf("config.New returned error: %s", err)
	}

	if viper.GetString("AVP_TYPE") != "awssecretsmanager" {
		t.Errorf("expected AVP_TYPE to be set from external config, was instead: %s", viper.GetString("AVP_TYPE"))
	}

	for envVar, expected := range expectedEnvVars {
		if actual := os.Getenv(envVar); actual != expected {
			t.Errorf("expected %s env to be %s, was instead: %s", envVar, expected, actual)
		}
	}

	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	os.Unsetenv("AWS_REGION")
}
