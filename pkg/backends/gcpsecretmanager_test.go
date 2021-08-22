package backends_test

import (
	"github.com/IBM/argocd-vault-plugin/pkg/backends"
	"github.com/googleapis/gax-go/v2"
	"golang.org/x/net/context"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"reflect"
	"testing"
)

type mockSecretManagerClient struct {
}

func (m *mockSecretManagerClient) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return &secretmanagerpb.AccessSecretVersionResponse{
		Name: "projects/project/secrets/test-secret/versions/2",
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte("some-value"),
		},
	}, nil
}

func TestGCPSecretManagerGetSecrets(t *testing.T) {
	ctx := context.Background()
	sm := backends.NewGCPSecretManagerBackend(ctx, &mockSecretManagerClient{})
	data, err := sm.GetSecrets("projects/project/secrets/test-secret/versions/2", map[string]string{})
	if err != nil {
		t.Fatalf("expected 0 errors but got: %s", err)
	}

	expected := map[string]interface{}{
		"test-secret": []byte("some-value"),
	}

	if !reflect.DeepEqual(expected, data) {
		t.Errorf("expected: %s, got: %s.", expected, data)
	}

}
