package backends

import (
	"context"
	"fmt"
	"github.com/googleapis/gax-go/v2"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"strings"
)

type SecretManagerClient interface {
	AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
}

// GCPSecretManager is a struct for working with a GCP Secret Manager backend
type GCPSecretManager struct {
	Context context.Context
	Client  SecretManagerClient
}

// NewGCPSecretManagerBackend initializes a new GCP Secret Manager backend
func NewGCPSecretManagerBackend(ctx context.Context, client SecretManagerClient) *GCPSecretManager {
	return &GCPSecretManager{
		Context: ctx,
		Client:  client,
	}
}

// Login does nothing as a "login" is handled on the instantiation of the Google SDK
func (a *GCPSecretManager) Login() error {
	return nil
}

// GetSecrets gets secrets from GCP Secret Manager and returns the formatted data
func (a *GCPSecretManager) GetSecrets(path string, _ map[string]string) (map[string]interface{}, error) {
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: path,
	}

	result, err := a.Client.AccessSecretVersion(a.Context, req)
	if err != nil {
		return nil, fmt.Errorf("Could not find secret: %v", err)
	}

	data := make(map[string]interface{})

	secretName := strings.Split(path, "/")[3]
	secretData := result.Payload.Data
	data[secretName] = secretData

	return data, nil
}
