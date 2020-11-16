package vault

// VaultType is an interface for the types of Vaults that are supported
type VaultType interface {
	Login() error
	GetSecrets(string) (map[string]interface{}, error)
}
