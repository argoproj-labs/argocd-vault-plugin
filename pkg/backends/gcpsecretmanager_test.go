package backends_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/backends"
	"github.com/googleapis/gax-go/v2"
	"golang.org/x/net/context"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

type mockSecretManagerClient struct {
	AccessSecretRequestName string
}

func (m *mockSecretManagerClient) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	m.AccessSecretRequestName = req.Name
	if strings.Contains(req.Name, "v3") {
		return &secretmanagerpb.AccessSecretVersionResponse{
			Name: "projects/project/secrets/test-secret/versions/v3",
			Payload: &secretmanagerpb.SecretPayload{
				Data: []byte("v3-value"),
			},
		}, nil
	}
	return &secretmanagerpb.AccessSecretVersionResponse{
		Name: "projects/project/secrets/test-secret/versions/2",
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte("some-value"),
		},
	}, nil
}

func TestGCPSecretManagerGetSecrets(t *testing.T) {
	ctx := context.Background()
	mock := mockSecretManagerClient{}
	sm := backends.NewGCPSecretManagerBackend(ctx, &mock)

	t.Run("GCP retrieve secrets", func(t *testing.T) {
		data, err := sm.GetSecrets("projects/project/secrets/test-secret", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		// Called correctly
		expectedCalledWith := "projects/project/secrets/test-secret/versions/latest"
		if !reflect.DeepEqual(expectedCalledWith, mock.AccessSecretRequestName) {
			t.Errorf("expected: %s, got: %s.", expectedCalledWith, mock.AccessSecretRequestName)
		}

		// Data correct
		expected := map[string]interface{}{
			"test-secret": "some-value",
		}

		if !reflect.DeepEqual(expected, data) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}
	})

	t.Run("GCP GetIndividualSecret", func(t *testing.T) {
		secret, err := sm.GetIndividualSecret("projects/project/secrets/test-secret", "test-secret", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := "some-value"

		if !reflect.DeepEqual(expected, secret) {
			t.Errorf("expected: %s, got: %s", expected, secret)
		}
	})

	t.Run("GCP retrieve secrets at version", func(t *testing.T) {
		data, err := sm.GetSecrets("projects/project/secrets/test-secret", "v3", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		// Called correctly
		expectedCalledWith := "projects/project/secrets/test-secret/versions/v3"
		if !reflect.DeepEqual(expectedCalledWith, mock.AccessSecretRequestName) {
			t.Errorf("expected: %s, got: %s.", expectedCalledWith, mock.AccessSecretRequestName)
		}

		// Data correct
		expected := map[string]interface{}{
			"test-secret": "v3-value",
		}
		if !reflect.DeepEqual(expected, data) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}
	})

	t.Run("GCP handle malformed path", func(t *testing.T) {
		_, err := sm.GetSecrets("ibmcloud/arbitrary/secrets/groups/some-group", "", map[string]string{})
		if err == nil {
			t.Fatalf("expected error")
		}

		expectedErr := "Path is not in the correct format"
		if !strings.Contains(err.Error(), expectedErr) {
			t.Fatalf("Expected error to have %s but said %s", expectedErr, err)
		}
	})
}
