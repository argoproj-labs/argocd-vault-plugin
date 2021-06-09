package helpers

import (
	"net"
	"testing"

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
	core, keyShares, rootToken := vault.TestCoreUnsealed(t)
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
			map[string]interface{}{
				"id": "1",
			},
			map[string]interface{}{
				"id": "2",
			},
			map[string]interface{}{
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

// CreateTestAppRoleVault initializes a new test vault with AppRole and Kv v2
func CreateTestAppRoleVault(t *testing.T) (*vault.TestCluster, string, string) {
	t.Helper()

	coreConfig := &vault.CoreConfig{
		CredentialBackends: map[string]logical.Factory{
			"approle": credAppRole.Factory,
		},
	}

	cluster := vault.NewTestCluster(t, coreConfig, &vault.TestClusterOptions{
		HandlerFunc: http.Handler,
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

	_, err = client.Logical().Write("auth/approle/role/role1", map[string]interface{}{
		"bind_secret_id": "true",
		"period":         "300",
		"policies":       "approle-secret, approle-kv",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Logical().Write("secret/testing", map[string]interface{}{
		"name":             "test-name",
		"namespace":        "test-namespace",
		"version":          "1.0",
		"replicas":         "2",
		"tag":              "1.0",
		"secret-var-value": "dGVzdC1wYXNzd29yZA==",
		"secret-num":       "MQ==",
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

	_, err = client.Logical().Write("kv/data/bad_test", map[string]interface{}{
		"hello": "world",
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

type MockVault struct {
	GetSecretsCalled bool
	Data             map[string]interface{}
}

func (v *MockVault) Login() error {
	return nil
}
func (v *MockVault) LoadData(data map[string]interface{}) {
	v.Data = data
}
func (v *MockVault) GetSecrets(path string, annotations map[string]string) (map[string]interface{}, error) {
	v.GetSecretsCalled = true
	return v.Data, nil
}
