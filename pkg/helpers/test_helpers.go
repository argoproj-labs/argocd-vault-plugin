package helpers

import (
	"net"
	"strconv"
	"testing"

	"github.com/hashicorp/go-hclog"
	kv "github.com/hashicorp/vault-plugin-secrets-kv"
	"github.com/hashicorp/vault/api"
	credAppRole "github.com/hashicorp/vault/builtin/credential/approle"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/hashicorp/vault/vault"
)

// CreateTestVault initializes a test vault with kv v2
func CreateTestVault(t *testing.T) (net.Listener, *api.Client, string) {
	t.Helper()

	// Create an in-memory, unsealed core (the "backend", if you will).
	core, keyShares, rootToken := vault.TestCoreUnsealedWithConfig(t, &vault.CoreConfig{
		Logger: hclog.NewNullLogger(),
	})
	_ = keyShares

	// Start an HTTP server for the core.
	ln, addr := http.TestServer(t, core)

	// Create a client that talks to the server, initially authenticating with
	// the root token.
	conf := api.DefaultConfig()
	conf.Address = addr

	client, err := api.NewClient(conf)
	if err != nil {
		t.Fatal(err)
	}
	client.SetToken(rootToken)

	client.Sys().Mount("kv", &api.MountInput{
		Type: "kv",
		Options: map[string]string{
			"version": "2",
		},
	})

	// Setup required secrets, policies, etc.
	_, err = client.Logical().Write("secret/foo", map[string]interface{}{
		"secret": "bar",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Logical().Write("kv/data/test", map[string]interface{}{
		"data": map[string]interface{}{
			"hello": "world",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Setup required secrets, policies, etc.
	_, err = client.Logical().Write("secret/ibm/arbitrary/groups/1", map[string]interface{}{
		"secrets": []map[string]interface{}{
			{
				"id": "1",
			},
			{
				"id": "2",
			},
			{
				"id": "3",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Setup required secrets, policies, etc.
	_, err = client.Logical().Write("secret/ibm/arbitrary/groups/1/1", map[string]interface{}{
		"name": "secret",
		"secret_data": map[string]interface{}{
			"payload": "value",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Setup required secrets, policies, etc.
	_, err = client.Logical().Write("secret/ibm/arbitrary/groups/1/2", map[string]interface{}{
		"name": "secret2",
		"secret_data": map[string]interface{}{
			"payload": "value2",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	return ln, client, rootToken
}

// CreateTestAppRoleVault initializes a new test vault with AppRole, database and Kv v2
func CreateTestAppRoleVault(t *testing.T) (*vault.TestCluster, string, string) {
	t.Helper()

	coreConfig := &vault.CoreConfig{
		LogicalBackends: map[string]logical.Factory{
			"kv": kv.Factory,
		},
		CredentialBackends: map[string]logical.Factory{
			"approle": credAppRole.Factory,
		},
	}

	cluster := vault.NewTestCluster(t, coreConfig, &vault.TestClusterOptions{
		HandlerFunc: http.Handler,
		Logger:      hclog.NewNullLogger(),
	})

	cluster.Start()

	vault.TestWaitActive(t, cluster.Cores[0].Core)

	client := cluster.Cores[0].Client

	client.Sys().Mount("kv", &api.MountInput{
		Type: "kv",
		Options: map[string]string{
			"version": "2",
		},
	})

	client.Sys().Mount("database", &api.MountInput{
		Type: "database",
	})

	if err := client.Sys().EnableAuthWithOptions("approle", &api.EnableAuthOptions{
		Type: "approle",
	}); err != nil {
		t.Fatal(err)
	}

	// Create Policy for secret/foo
	err := client.Sys().PutPolicy("approle-secret", "path \"secret/*\" { capabilities = [\"read\",\"list\"] }")
	if err != nil {
		t.Fatal(err)
	}

	// Create Policy for kv
	err = client.Sys().PutPolicy("approle-kv", "path \"kv/*\" { capabilities = [\"read\",\"list\"] }")
	if err != nil {
		t.Fatal(err)
	}

	// Create Policy for database
	err = client.Sys().PutPolicy("approle-database", "path \"database/*\" { capabilities = [\"read\",\"list\"] }")
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Logical().Write("auth/approle/role/role1", map[string]interface{}{
		"bind_secret_id": "true",
		"period":         "300",
		"policies":       "approle-secret, approle-kv, approle-database",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Logical().Write("secret/testing", map[string]interface{}{
		"name":              "test-name",
		"namespace":         "test-namespace",
		"version":           "1.0",
		"replicas":          "2",
		"tag":               "1.0",
		"secret-var-value":  "dGVzdC1wYXNzd29yZA==",
		"secret-var-value2": "dGVzdC1wYXNzd29yZDI=",
		"secret-num":        "MQ==",
		"secret-var-clear":  "test-password",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Logical().Write("kv/data/testing", map[string]interface{}{
		"data": map[string]interface{}{
			"name":        "test-kv-name",
			"namespace":   "test-kv-namespace",
			"version":     "1.2",
			"replicas":    "3",
			"tag":         "1.1",
			"target-port": 80,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Logical().Write("secret/foo", map[string]interface{}{
		"secret": "bar",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Logical().Write("secret/json", map[string]interface{}{
		"data": map[string]interface{}{
			"service": map[string]interface{}{
				"enableTLS": true,
				"ports":     []int{80, 8080},
			},
			"deployment": map[string]interface{}{
				"replicas": 2,
				"image": map[string]interface{}{
					"name": "json-test",
					"tag":  "latest",
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Logical().Write("secret/jsonstring", map[string]interface{}{
		"secret": "{\"credentials\":{\"user\":\"test-user\",\"pass\":\"test-password\"}}",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Logical().Write("kv/data/test", map[string]interface{}{
		"data": map[string]interface{}{
			"hello": "world",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Logical().Write("secret/bad_test", map[string]interface{}{
		"hello": "world",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Logical().Write("kv/data/versioned", map[string]interface{}{
		"data": map[string]interface{}{
			"secret": "version1",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Logical().Write("kv/data/versioned", map[string]interface{}{
		"data": map[string]interface{}{
			"secret": "version2",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Logical().Write("secret/base64", map[string]interface{}{
		"encoded_secret": "ewogICJrZXkxIjogInNlY3JldDEiLAogICJrZXkyIjogInNlY3JldDIiLAogICJrZXkzIjogInNlY3JldDMiCn0K",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Logical().Write("secret/yaml", map[string]interface{}{
		"secret": "---\nkey1: secret1\nkey2: secret2\nkey3: secret3",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Logical().Write("database/config/testing", map[string]interface{}{
		"plugin_name": "postgresql-database-plugin",
		"allowed_roles": "testing*",
		"connection_url": "TODO",
		"username": "TODO",
		"password": "TODO",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Logical().Write("database/static-roles/testing", map[string]interface{}{
		"db_name": "testing",
		"username": "testing",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Logical().Write("database/roles/testing1", map[string]interface{}{
		"db_name": "testing",
		"creation_statements": "TODO",
	})
	if err != nil {
		t.Fatal(err)
	}

	secret, err := client.Logical().Write("auth/approle/role/role1/secret-id", nil)
	if err != nil {
		t.Fatal(err)
	}
	secretID := secret.Data["secret_id"].(string)

	secret, err = client.Logical().Read("auth/approle/role/role1/role-id")
	if err != nil {
		t.Fatal(err)
	}
	roleID := secret.Data["role_id"].(string)

	return cluster, roleID, secretID
}

// CreateTestGithubVault initializes a new test vault with AppRole and Kv v2
func CreateTestAuthVault(t *testing.T) *vault.TestCluster {
	t.Helper()

	coreConfig := &vault.CoreConfig{
		CredentialBackends: map[string]logical.Factory{
			"github":     Factory,
			"kubernetes": Factory,
			"ibmcloud":   Factory,
		},
	}

	cluster := vault.NewTestCluster(t, coreConfig, &vault.TestClusterOptions{
		HandlerFunc: http.Handler,
		Logger:      hclog.NewNullLogger(),
	})

	cluster.Start()

	vault.TestWaitActive(t, cluster.Cores[0].Core)

	client := cluster.Cores[0].Client

	client.Sys().Mount("kv", &api.MountInput{
		Type: "kv",
		Options: map[string]string{
			"version": "2",
		},
	})

	if err := client.Sys().EnableAuthWithOptions("github", &api.EnableAuthOptions{
		Type: "github",
	}); err != nil {
		t.Fatal(err)
	}

	if err := client.Sys().EnableAuthWithOptions("kubernetes", &api.EnableAuthOptions{
		Type: "kubernetes",
	}); err != nil {
		t.Fatal(err)
	}

	if err := client.Sys().EnableAuthWithOptions("ibmcloud", &api.EnableAuthOptions{
		Type: "ibmcloud",
	}); err != nil {
		t.Fatal(err)
	}

	return cluster
}

// MockVault is used to mock out a generic SM Backend
// It's useful for testing replacement behavior
type MockVault struct {
	GetSecretsCalled          bool
	GetIndividualSecretCalled bool
	Data                      []map[string]interface{}
}

func (v *MockVault) Login() error {
	return nil
}
func (v *MockVault) LoadData(data map[string]interface{}) {
	v.Data = append(v.Data, data)
}
func (v *MockVault) GetSecrets(path string, version string, annotations map[string]string) (map[string]interface{}, error) {
	v.GetSecretsCalled = true
	if len(v.Data) == 0 {
		return make(map[string]interface{}), nil
	}
	if version == "" {
		return v.Data[len(v.Data)-1], nil
	}
	num, _ := strconv.ParseInt(version, 10, 0)
	return v.Data[num-1], nil
}
func (v *MockVault) GetIndividualSecret(path, secret, version string, annotations map[string]string) (interface{}, error) {
	v.GetIndividualSecretCalled = true
	if len(v.Data) == 0 {
		return nil, nil
	}
	if version == "" {
		return v.Data[len(v.Data)-1][secret], nil
	}
	num, _ := strconv.ParseInt(version, 10, 0)
	return v.Data[num-1][secret], nil
}
