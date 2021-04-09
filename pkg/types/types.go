package types

import (
	"github.com/hashicorp/vault/api"
)

// Backend is an interface for the types of Vaults that are supported
type Backend interface {
	Login() error
	GetSecrets(string, map[string]string) (map[string]interface{}, error)
}

// AuthType is and interface for the supported authentication methods
type AuthType interface {
	Authenticate(*api.Client) error
}
