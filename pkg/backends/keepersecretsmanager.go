package backends

import (
	"fmt"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	ksm "github.com/keeper-security/secrets-manager-go/core"
)

// KeeperClient is an interface containing the aspects of the keeper api that the backend needs.
type KeeperClient interface {
	GetSecrets(ids []string) ([]*ksm.Record, error)
}

// KeeperSecretsManager is a struct for working with a Keeper Secrets Manager backend
type KeeperSecretsManager struct {
	client KeeperClient
}

// NewKeeperSecretsManagerBackend returns a new Keeper Secrets Manager backend.
func NewKeeperSecretsManagerBackend(client KeeperClient) *KeeperSecretsManager {
	return &KeeperSecretsManager{
		client: client,
	}
}

// Login currently does nothing.
func (k *KeeperSecretsManager) Login() error {
	return nil
}

func buildSecretsMap(secretsMap map[string]interface{}, fieldMap map[string]interface{}) error {
	var label string
	var fieldType string

	// we start by getting the label from the type,
	if fieldMapType, ok := fieldMap["type"]; ok {
		if typeString, ok := fieldMapType.(string); ok {
			label = typeString
			fieldType = typeString
		}
	}

	// if there is a more specific label, then we use that instead
	if labelField, ok := fieldMap["label"]; ok {
		if labelStr, ok := labelField.(string); ok {
			label = labelStr
		}
	}

	if label == "" {
		utils.VerboseToStdErr("secret's field does not have a \"label\"")
		return nil
	}

	if fieldType == "" {
		utils.VerboseToStdErr("secret's field does not have a \"type\"")
		return nil
	}

	var value interface{}

	switch fieldType {
	case "note":
		fallthrough
	case "url":
		fallthrough
	case "text":
		fallthrough
	case "login":
		fallthrough
	case "password":
		fallthrough
	default:
		if fieldValues, ok := fieldMap["value"]; ok {
			sliceValues, ok := fieldValues.([]interface{})
			if !ok {
				utils.VerboseToStdErr("secret's field value is not a []interface %T for field %s", fieldValues, label)
				return nil
			}

			if len(sliceValues) == 1 {
				value = sliceValues[0]
			}

			if len(sliceValues) == 0 {
				value = nil
			}
		}
	}

	if value != nil {
		secretsMap[label] = value
	}
	return nil
}

// GetSecrets gets secrets from Keeper Secrets Manager. It does not currently
// implement anything related to versions or annotations.
func (a *KeeperSecretsManager) GetSecrets(path string, version string, annotations map[string]string) (map[string]interface{}, error) {
	records, err := a.client.GetSecrets([]string{
		path,
	})
	if err != nil {
		return nil, fmt.Errorf("could not access secret %s, error: %s", path, err)
	}
	utils.VerboseToStdErr("Keeper Secrets Manager getting path %s", path)

	if len(records) == 0 {
		return nil, fmt.Errorf("no secrets could be found with the given path: %s", path)
	}

	if len(records) > 1 {
		return nil, fmt.Errorf("unexpectedly multiple secrets were found with the given path: %s", path)
	}

	utils.VerboseToStdErr("Keeper Secrets Manager decoding record %s", records[0].Title())

	dict := records[0].RecordDict

	secretMap := map[string]interface{}{}

	if fields, ok := dict["custom"]; ok {
		if fieldsSlice, ok := fields.([]interface{}); ok {
			for _, field := range fieldsSlice {
				if fieldMap, ok := field.(map[string]interface{}); ok {
					buildSecretsMap(secretMap, fieldMap)
				}
			}
		}
	}

	if fields, ok := dict["fields"]; ok {
		if fieldsSlice, ok := fields.([]interface{}); ok {
			for _, field := range fieldsSlice {
				if fieldMap, ok := field.(map[string]interface{}); ok {
					buildSecretsMap(secretMap, fieldMap)
				}
			}
		}
	}

	utils.VerboseToStdErr("Keeper Secrets Manager constructed map %s", secretMap)

	return secretMap, nil

}

// GetIndividualSecret returns the specified secret. It simply wraps the
// GetSecrets call, and currently ignores the version parameter.
func (v *KeeperSecretsManager) GetIndividualSecret(kvpath, secretName, version string, annotations map[string]string) (interface{}, error) {
	secrets, err := v.GetSecrets(kvpath, version, annotations)
	if err != nil {
		return nil, err
	}

	return secrets[secretName], nil
}
