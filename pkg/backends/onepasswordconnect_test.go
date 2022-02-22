package backends_test

import (
	"reflect"
	"testing"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/backends"
)

type mockOnePasswordConnectClient struct {
	connect.Client
}

func (m *mockOnePasswordConnectClient) GetItem(uuid, vaultUUID string) (*onepassword.Item, error) {
	data := &onepassword.Item{}

	switch uuid {
	case "test":
		data.Fields = []*onepassword.ItemField{
			{
				ID:       "",
				Section:  nil,
				Type:     "",
				Purpose:  "",
				Label:    "test-secret",
				Value:    "current-value",
				Generate: false,
				Recipe:   nil,
				Entropy:  0,
			},
		}
	}

	return data, nil
}

func TestOnePasswordConnectGetSecrets(t *testing.T) {
	sm := backends.NewOnePasswordConnectBackend(&mockOnePasswordConnectClient{})

	t.Run("Get secrets", func(t *testing.T) {
		data, err := sm.GetSecrets("vaults/vault1/items/test", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"test-secret": "current-value",
		}

		if !reflect.DeepEqual(expected, data) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}
	})

	t.Run("1Password GetIndividualSecret", func(t *testing.T) {
		secret, err := sm.GetIndividualSecret("vaults/vault1/items/test", "test-secret", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := "current-value"

		if !reflect.DeepEqual(expected, secret) {
			t.Errorf("expected: %s, got: %s.", expected, secret)
		}
	})
}
