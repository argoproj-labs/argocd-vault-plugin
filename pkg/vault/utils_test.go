package vault

import (
	"net"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/vault"
)

func CreateTestVault(t *testing.T) (net.Listener, *api.Client) {
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

	// Setup required secrets, policies, etc.
	_, err = client.Logical().Write("secret/foo", map[string]interface{}{
		"secret": "bar",
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

	return ln, client
}
