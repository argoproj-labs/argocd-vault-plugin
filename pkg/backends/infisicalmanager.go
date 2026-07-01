package backends

import (
	"fmt"

	infisical "github.com/infisical/go-sdk"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
)

type InfisicalSecretsIface interface {
	List(options infisical.ListSecretsOptions) ([]infisical.Secret, error)
	Retrieve(options infisical.RetrieveSecretOptions) (infisical.Secret, error)
}

// Infisical is a struct for working with an Infisical backend
type Infisical struct {
	client      InfisicalSecretsIface
	projectSlug string
	environment string
}

// NewInfisicalBackend initializes a new Infisical backend
func NewInfisicalBackend(client InfisicalSecretsIface, projectSlug, environment string) *Infisical {
	return &Infisical{
		client:      client,
		projectSlug: projectSlug,
		environment: environment,
	}
}

// Login does nothing since authentication is handled during client initialization
func (i *Infisical) Login() error {
	return nil
}

// GetSecrets gets secrets from Infisical and returns the formatted data
func (i *Infisical) GetSecrets(path string, version string, annotations map[string]string) (map[string]interface{}, error) {
	utils.VerboseToStdErr("Infisical listing secrets at path %s", path)

	if path == "" {
		path = "/"
	}

	secrets, err := i.client.List(infisical.ListSecretsOptions{
		ProjectSlug:            i.projectSlug,
		Environment:            i.environment,
		SecretPath:             path,
		ExpandSecretReferences: true,
		IncludeImports:         true,
	})
	if err != nil {
		return nil, fmt.Errorf("could not list Infisical secrets at path %s: %w", path, err)
	}

	result := make(map[string]interface{}, len(secrets))
	for _, s := range secrets {
		result[s.SecretKey] = s.SecretValue
	}

	utils.VerboseToStdErr("Infisical list secrets response: %v", result)
	return result, nil
}

// GetIndividualSecret retrieves a specific secret from Infisical
func (i *Infisical) GetIndividualSecret(path, secret, version string, annotations map[string]string) (interface{}, error) {
	utils.VerboseToStdErr("Infisical getting secret %s at path %s", secret, path)

	if path == "" {
		path = "/"
	}

	s, err := i.client.Retrieve(infisical.RetrieveSecretOptions{
		SecretKey:              secret,
		ProjectSlug:            i.projectSlug,
		Environment:            i.environment,
		SecretPath:             path,
		IncludeImports:         true,
		ExpandSecretReferences: true,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get Infisical secret %s at path %s: %w", secret, path, err)
	}

	return s.SecretValue, nil
}
