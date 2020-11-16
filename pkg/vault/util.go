package vault

import (
	"fmt"
	"os"
)

// InitVault TODO
func InitVault(vaultType string) (VaultType, error) {
	switch vaultType {
	case "Vault":
		{
			return &Github{
				AccessToken: os.Getenv("AVP_GITHUB_TOKEN"),
			}, nil
		}
	case "SecretManager":
		{
			return &SecretManager{
				IAMToken: os.Getenv("AVP_IAM_TOKEN"),
			}, nil
		}
	}
	return nil, fmt.Errorf("unsupported Vault: %s", vaultType)
}
