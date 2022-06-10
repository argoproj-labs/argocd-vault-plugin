package backends

import (
	"encoding/json"
	"fmt"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
)

// AWSSecretsManager is a struct for working with a AWS Secrets Manager backend
type AWSSecretsManager struct {
	Client secretsmanageriface.SecretsManagerAPI
}

// NewAWSSecretsManagerBackend initializes a new AWS Secrets Manager backend
func NewAWSSecretsManagerBackend(client secretsmanageriface.SecretsManagerAPI) *AWSSecretsManager {
	return &AWSSecretsManager{
		Client: client,
	}
}

// Login does nothing as a "login" is handled on the instantiation of the aws sdk
func (a *AWSSecretsManager) Login() error {
	return nil
}

// GetSecrets gets secrets from aws secrets manager and returns the formatted data
func (a *AWSSecretsManager) GetSecrets(path string, version string, annotations map[string]string) (map[string]interface{}, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(path),
	}

	if version != "" {
		input.SetVersionId(version)
	}

	utils.VerboseToStdErr("AWS Secrets Manager getting secret %s at version %s", path, version)
	result, err := a.Client.GetSecretValue(input)
	if err != nil {
		return nil, err
	}

	utils.VerboseToStdErr("AWS Secrets Manager get secret response %v", result)

	var dat map[string]interface{}

	if result.SecretString != nil {
		err := json.Unmarshal([]byte(*result.SecretString), &dat)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("Could not find secret %s", path)
	}

	return dat, nil
}

// GetIndividualSecret will get the specific secret (placeholder) from the SM backend
// For AWS, we only support placeholders replaced from the k/v pairs of a secret which cannot be individually addressed
// So, we use GetSecrets and extract the specific placeholder we want
func (a *AWSSecretsManager) GetIndividualSecret(kvpath, secret, version string, annotations map[string]string) (interface{}, error) {
	data, err := a.GetSecrets(kvpath, version, annotations)
	if err != nil {
		return nil, err
	}
	return data[secret], nil
}
