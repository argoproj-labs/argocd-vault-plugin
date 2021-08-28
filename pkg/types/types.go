package types

import (
	"net/http"

	"github.com/hashicorp/vault/api"
)

// Backend is an interface for the types of Vaults that are supported
type Backend interface {
	Login() error

	// GetSecrets retrieves the secret at `path` with specified `version` based on configuation given in `annotations`
	GetSecrets(path string, version string, annotations map[string]string) (map[string]interface{}, error)
}

// AuthType is and interface for the supported authentication methods
type AuthType interface {
	Authenticate(*api.Client) error
}

// HTTPClient interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
