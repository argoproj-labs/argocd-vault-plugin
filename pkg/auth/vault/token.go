package vault

import (
	"github.com/hashicorp/vault/api"
)

// Just a plain vault token
type TokenAuth struct{}

// The vault client auto-detect if the VAULT_TOKEN is set, so
// just leave everything to the vault client
func (t *TokenAuth) Authenticate(vaultClient *api.Client) error {
	return nil
}
