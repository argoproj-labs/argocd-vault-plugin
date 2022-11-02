package backends

import (
	"context"
	"fmt"
	"regexp"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/types"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"github.com/googleapis/gax-go/v2"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

var GCPPath, _ = regexp.Compile(`projects/(?P<projectid>.+)/secrets/(?P<secretid>.+)`)

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
// The path is of format `projects/project-id/secrets/secret-id`
func (a *GCPSecretManager) GetSecrets(path string, version string, annotations map[string]string) (map[string]interface{}, error) {
	matches := GCPPath.FindStringSubmatch(path)
	if len(matches) == 0 {
		return nil, fmt.Errorf("Path is not in the correct format (projects/$PROJECT_ID/secrets/$SECRET_ID) for GCP Secrets Manager: %s", path)
	}

	if version == "" {
		version = types.GCPCurrentSecretVersion
	}
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("%s/versions/%s", path, version),
	}

	utils.VerboseToStdErr("GCP Secret Manager accessing secret at path %s at version  %v", path, version)
	result, err := a.Client.AccessSecretVersion(a.Context, req)
	if err != nil {
		return nil, fmt.Errorf("Could not find secret: %v", err)
	}

	utils.VerboseToStdErr("GCP Secret Manager access secret version response %v", result)

	data := make(map[string]interface{})

	secretName := matches[GCPPath.SubexpIndex("secretid")]
	secretData := result.Payload.Data
	data[secretName] = string(secretData)

	return data, nil
}

// GetIndividualSecret will get the specific secret (placeholder) from the SM backend
// For GCP, the path is specific to the secret
// So, we just forward the value from the k/v result of GetSecrets
func (a *GCPSecretManager) GetIndividualSecret(kvpath, secret, version string, annotations map[string]string) (interface{}, error) {
	data, err := a.GetSecrets(kvpath, version, annotations)
	if err != nil {
		return nil, err
	}
	return data[secret], nil
}
