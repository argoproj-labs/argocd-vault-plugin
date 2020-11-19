package vault

import (
	"reflect"
	"testing"
)

func TestGithubGetSecrets(t *testing.T) {
	ln, client := CreateTestVault(t)
	defer ln.Close()

	vc := &Client{
		PathPrefix:     "secret",
		VaultAPIClient: client,
	}

	github := Github{
		AccessToken: "token",
		Client:      vc,
	}

	expected := map[string]interface{}{
		"secret": "bar",
	}

	data, err := github.GetSecrets("/foo")
	if err != nil {
		t.Fatalf("expected 0 errors but got: %s", err)
	}

	if !reflect.DeepEqual(data, expected) {
		t.Errorf("expected: %s, got: %s.", expected, data)
	}
}
