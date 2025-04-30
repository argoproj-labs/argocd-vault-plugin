package config_test

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/config"
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
				"AVP_TYPE":           "vault",
				"AVP_AUTH_TYPE":      "k8s",
				"AVP_K8S_MOUNT_PATH": "mount_point",
				"AVP_K8S_ROLE":       "role",
				"AVP_K8S_TOKEN_PATH": "toke_path",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":      "vault\n",
				"AVP_AUTH_TYPE": "k8s",
				"AVP_K8S_ROLE":  "role",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":           "vault",
				"AVP_AUTH_TYPE":      "k8s",
				"AVP_K8S_MOUNT_PATH": "mount_point",
				"AVP_K8S_ROLE":       "role",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":       "vault",
				"AVP_AUTH_TYPE":  "k8s",
				"AVP_MOUNT_PATH": "mount_point",
				"AVP_K8S_ROLE":   "role",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":      "vault",
				"AVP_AUTH_TYPE": "userpass",
				"AVP_USERNAME":  "username",
				"AVP_PASSWORD":  "password",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":       "vault",
				"AVP_AUTH_TYPE":  "userpass",
				"AVP_MOUNT_PATH": "mount_path",
				"AVP_USERNAME":   "username",
				"AVP_PASSWORD":   "password",
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
		{
			map[string]interface{}{
				"AVP_TYPE":                   "yandexcloudlockbox",
				"AVP_YCL_KEY_ID":             "test",
				"AVP_YCL_SERVICE_ACCOUNT_ID": "test",
				"AVP_YCL_PRIVATE_KEY": `-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQCIwbOQ4mB4LlFKNvvkot8qnKoffHLxVu2+DNpKC3WiPbof23bf
eHcFTj14/h3HP75dxH5GIop2C8HQzyGzScIEHMxqOpwgu8+tmHbsCAdWkbC03wQ0
1++nHmI6kAUx0mFDAXGovyDiR132iZ5lX2hEJ2Nd2g67SHV140sB6T0vRQIDAQAB
AoGASx2B4NnGvRxLwCTVVK71PzWP5/12MQNbUGFE4RjMQxH+kpL8ByDm1v4zm6qQ
dqmXiW9tIF7GiLJKgcPTseOYcdQkGlST1MgYAqtxMkGYYCP94cGna0qy4lIFBJee
B/dKY56UiIEtJbMvN/T9LFBx1Kw5jT4R5lhdysuabsqAt+ECQQDYHmfMee/Dzw+/
G4xlJfIfcQ4648/zf53hlA5MwCBbm6wv2KLkWglzSl9Vy54f/UM4VtIfywjmTkj+
C2b17Uq9AkEAof4tJwllt4AwIjIp1KEiBTY6z0Whoe9SO5RqFmBUkVTeiIuUxgGE
+NLCY+0NzG2FNglT96ik/Xxi+/uiy4wDKQJBAIQ9TpwyfIBe4a65R5XYuyd8AQ4N
uX+wNcYC1yElamdDgP+h2kJJyYCPIHiZ5/6A9LGzhk1H6gEqI8W26mBOuy0CQEcl
y88JYZNmyb07KwQogTioyMugWY01/3gLh0ysonfyPoraQ01z/WMLrjUVOKpAr/E7
x5VOjKiIqTDjJG0h4YECQQDR7tTAXzccGQmHhmN72mDB5LfWi8uSADT4gsimY82m
fDGt+yaf3RaZbVwHSVLzxiXGsu1WQJde3uJeNh5c6z+5
-----END RSA PRIVATE KEY-----`,
			},
			"*backends.YandexCloudLockbox",
		},
		{
			map[string]interface{}{
				"AVP_TYPE": "sops",
			},
			"*backends.LocalSecretManager",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":         "1passwordconnect",
				"OP_CONNECT_TOKEN": "token",
				"OP_CONNECT_HOST":  "opconnect.somedomain.com",
			},
			"*backends.OnePasswordConnect",
		},
		{
			map[string]interface{}{
				"ARGOCD_ENV_AVP_TYPE":         "vault",
				"ARGOCD_ENV_AVP_AUTH_TYPE":    "github",
				"ARGOCD_ENV_AVP_GITHUB_TOKEN": "token",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"ARGOCD_ENV_AVP_TYPE":         "vault",
				"AVP_TYPE":                    "not-valid-type",
				"ARGOCD_ENV_AVP_AUTH_TYPE":    "github",
				"ARGOCD_ENV_AVP_GITHUB_TOKEN": "token",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":               "keepersecretsmanager",
				"AVP_KEEPER_CONFIG_PATH": "/mnt/foobar",
			},
			"*backends.KeeperSecretsManager",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":             "delineasecretserver",
				"AVP_AUTH_TYPE":        "userpass",
				"AVP_DELINEA_URL":      "http://my-delinea-server",
				"AVP_DELINEA_USER":     "username",
				"AVP_DELINEA_PASSWORD": "password",
			},
			"*backends.DelineaSecretServer",
		},
		{
			map[string]interface{}{
				"AVP_TYPE": "kubernetessecret",
			},
			"*backends.KubernetesSecret",
		},
		{
			map[string]interface{}{
				"AVP_TYPE": "ocivault",
				"AVP_OCI_TENANCY": "ocid1.tenancy.oc1..aaaaaaaa",
				"AVP_OCI_USER": "ocid1.user.oc1..aaaaaaaa",
				"AVP_OCI_REGION": "test-region",
				"AVP_OCI_FINGERPRINT": "test-fingerprint",
				"AVP_OCI_VAULT_ID": "test-vault-id",
				"AVP_OCI_VAULT_COMPARTMENT_ID": "test-vault-comp-id",
				"AVP_OCI_KEY_FILE": `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCyDz0+WvWGmcym
OEBQ0zhWO1Abs/UQ1v0A7kXQpTgwAFKO0SR56jJBII1VmBuctDUYkdO55FvAuhNv
WTQcxi0WjVejZ7npNhyU1vcDVmJGmVc1vgb6Jo+pfDGq/EnkaNthpE50+jPeA3jD
l6/NHaHb4f9UUpX7o/YfG8RQ1kF7Dkq1YqVEPBrRfZOzenIE0CQjRjR6GnwdAePB
GtrmHP3ZOM/t5QlKYd68z2xQfbvl88/W99PnjKMLmhyObWzwtcpWsjnjLm8vg49X
ZOa7+y14/Ev/FR7HOAwsXpXFGtPiib5JcoFIddc+bdjlSaINx3nOq9lal/nNHv8Q
hcSMg+RxAgMBAAECggEAV3EUamLQ4GD3F0nYi9iueep21KPzXWm2pZZdwrDgfvIp
mOksOJLCSylpPveL19DHomE60LdMN8Epei0cYmUQD1sqBp0Rt21Ta+SFOaZabMEx
CrtfQcleE6Vh3s42m2zDD5hYzylv/z9FNwhu1RQQQKMjeI12CjXi0DQanHgbgAoZ
GN73kAAtwoi/wClpJsbMypshkShQGdMzBaXOh8rm5NGCOuzM0L2Ovm1yZdfn9in1
T1ivStRXojAkWFYJ2PATpR0USEFboKahUDWHRMUUQhHXY5M07G9mxmyj1yeT3bPk
yD4cFW8PHLRhfiM4FAX5/0ATpeaS+9yjRaZgqRknYQKBgQD0ZA9TyOe/iYFIXeuZ
7fi/V08sJ7ULR+xRj65KYgOlowsRl4EzSURgaprMPpJQ4B6bWt4VHcxX3rpAnTie
46LgOqU1iK9LYTBgvTpgnhLxzvKYosgtE8qrPnoiAmtXR+cBEcHTXFlomdwFBO40
tNZaLQi9WpcZvw87t5mm6t5E6wKBgQC6hIznWSj51fcO3+OobQCZql3lSftJ/6EV
BJ576uuaI4fbYgrw8O7q2WULVLeS34KERYS5mxQUhnkaC5hwQzYRVJ2lbsBFMqhv
pT9maKSEqMfDm5iJmSz0k+L4eJaU7xR/RMaZRtQLR7MfY9nyKGeP97e0Kk1rhYF7
7oNS5biVEwKBgQDq3PY16M9+rSDHcSsoRSBWkguOPaKpcrdTMqem6Ebk+al7gIQz
y2eg2RJm0oM+ogQH/O2MkZR9pZiM3As79zviDboTloYQBRi+/1uI2qEOLXnK4jVJ
zMlqhKJO6NBLktgXmP8Spp9t/N8LG8/oaxnMk5bgkpy/q3NySmGpnfF5fQKBgGkC
69ntBvbyknCbeS+Af1AE7WyEpKha9jRBL4GRGCjmTD0mDAbvf3RWBV/FyL02feM+
yKU/PKT5uQEC+kZqcOx8+W0E19ed19tT7EgaLlZKOH5XAiCmTvs8sBM4wX8ExEOL
U01E5WmcaqsHqtN+ECCsVY9oKcKZnfdKqEFp+OxlAoGAMrCGEePfP3z6uLQW5JZ6
/F3yvIIbFf9qERwr56eyk6MPqZMN+QdZuVVY0qFxj0F1NRKFJN2IZU8LPnmNICuh
Ap0WFIpwv5OJ94OlooATeYakqTj5l2gdQkgxAI8bZlqO7XVtVEPSvXCoLrZKEVqA
Zfzm2fgQCYKEhd8HQDmkT/8=
-----END PRIVATE KEY-----`,
			},
			"*backends.OCIVault",
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
	expectedError := "Must provide a supported Vault Type, received "

	if err.Error() != expectedError {
		t.Errorf("expected error %s to be thrown, got %s", expectedError, err)
	}
}

func TestNewConfigNoAuthType(t *testing.T) {
	os.Setenv("AVP_TYPE", "vault")
	viper := viper.New()
	_, err := config.New(viper, &config.Options{})
	expectedError := "Must provide a supported Authentication Type, received "

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
			"warning: AWS_REGION env var not set, using AWS region us-east-2\n",
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
		viper.Set("verboseOutput", true)

		v := viper.New()
		output := captureOutput(func() {
			config, err := config.New(v, &config.Options{})
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			xType := fmt.Sprintf("%T", config.Backend)
			if xType != tc.expectedType {
				t.Errorf("expected: %s, got: %s.", tc.expectedType, xType)
			}
		})

		if !strings.Contains(output, tc.expectedLog) {
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
				"AVP_TYPE":      "vault",
				"AVP_AUTH_TYPE": "userpass",
				"AVP_USERNAME":  "username",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":      "vault",
				"AVP_AUTH_TYPE": "userpass",
				"AVP_PASSWORD":  "password",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":      "vault",
				"AVP_AUTH_TYPE": "userpass",
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
				"AVP_TYPE":                   "yandexcloudlockbox",
				"AVP_YCL_KEY_ID":             "test",
				"AVP_YCL_SERVICE_ACCOUNT_ID": "test",
			},
			"*backends.YandexCloudLockbox",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":         "1passwordconnect",
				"OP_CONNECT_TOKEN": "token",
			},
			"*backends.OnePasswordConnect",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":         "delineasecretserver",
				"AVP_AUTH_TYPE":    "userpass",
				"AVP_DELINEA_URL":  "http://my-delinea-server",
				"AVP_DELINEA_USER": "username",
			},
			"*backends.DelineaSecretServer",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":             "delineasecretserver",
				"AVP_AUTH_TYPE":        "userpass",
				"AVP_DELINEA_URL":      "http://my-delinea-server",
				"AVP_DELINEA_PASSWORD": "password",
			},
			"*backends.DelineaSecretServer",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":             "delineasecretserver",
				"AVP_AUTH_TYPE":        "userpass",
				"AVP_DELINEA_USER":     "username",
				"AVP_DELINEA_PASSWORD": "password",
			},
			"*backends.DelineaSecretServer",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":         "delineasecretserver",
				"AVP_AUTH_TYPE":    "userpass",
				"AVP_DELINEA_USER": "username",
			},
			"*backends.DelineaSecretServer",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":             "delineasecretserver",
				"AVP_AUTH_TYPE":        "userpass",
				"AVP_DELINEA_PASSWORD": "password",
			},
			"*backends.DelineaSecretServer",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":      "delineasecretserver",
				"AVP_AUTH_TYPE": "userpass",
			},
			"*backends.DelineaSecretServer",
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
	tmpFile, err := os.CreateTemp("", "avpConfig.*.yaml")
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

func TestExternalConfigSOPS(t *testing.T) {
	const avpSOPSConfig = `AVP_TYPE: sops
SOPS_AGE_KEY_FILE: age`

	expectedSOPSEnvVars := map[string]string{
		"AVP_TYPE":          "", // shouldn't be an env var
		"SOPS_AGE_KEY_FILE": "age",
	}

	// Test setting SOPS_* env variables from external AVP config, note setting
	// env vars is necessary to pass AVP config entries to SOPS
	tmpFile, err := os.CreateTemp("", "avpSOPSConfig.*.yaml")
	if err != nil {
		t.Errorf("Cannot create temporary file %s", err)
	}

	defer os.Remove(tmpFile.Name()) // clean up the file afterwards

	if _, err = tmpFile.WriteString(avpSOPSConfig); err != nil {
		t.Errorf("Failed to write to temporary file %s", err)
	}

	viper := viper.New()
	if _, err = config.New(viper, &config.Options{ConfigPath: tmpFile.Name()}); err != nil {
		t.Errorf("config.New returned error: %s", err)
	}

	if viper.GetString("AVP_TYPE") != "sops" {
		t.Errorf("expected AVP_TYPE to be set from external config, was instead: %s", viper.GetString("AVP_TYPE"))
	}

	for envVar, expected := range expectedSOPSEnvVars {
		if actual := os.Getenv(envVar); actual != expected {
			t.Errorf("expected %s env to be %s, was instead: %s", envVar, expected, actual)
		}
	}

	os.Unsetenv("SOPS_AGE_KEY_FILE")
}
