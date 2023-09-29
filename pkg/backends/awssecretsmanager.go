package backends

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

const (
	AWS_CURRENT  string = "AWSCURRENT"
	AWS_PREVIOUS string = "AWSPREVIOUS"
)

type AWSSecretsManagerIface interface {
	GetSecretValue(ctx context.Context,
		params *secretsmanager.GetSecretValueInput,
		optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// AWSSecretsManager is a struct for working with a AWS Secrets Manager backend
type AWSSecretsManager struct {
	Client AWSSecretsManagerIface
}

// NewAWSSecretsManagerBackend initializes a new AWS Secrets Manager backend
func NewAWSSecretsManagerBackend(client AWSSecretsManagerIface) *AWSSecretsManager {
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
	var opts = func(o *secretsmanager.Options) {}

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(path),
	}

	re := regexp.MustCompile(`(?m)^(?:[^:]+:){3}([^:]+).*`)
	if re.MatchString(path) {
		parts := re.FindStringSubmatch(path)

		opts = func(o *secretsmanager.Options) {
			o.Region = parts[1]
		}
	}

	if version != "" {
		if version == AWS_CURRENT || version == AWS_PREVIOUS {
			input.VersionStage = aws.String(version)
		} else {
			input.VersionId = aws.String(version)
		}
	}

	utils.VerboseToStdErr("AWS Secrets Manager getting secret %s at version %s", path, version)
	result, err := a.Client.GetSecretValue(context.TODO(), input, opts)
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
	} else if result.SecretBinary != nil {
		utils.VerboseToStdErr("Get binary value for %v", path)
		dat = make(map[string]interface{})
		dat["SecretBinary"] = result.SecretBinary
		return dat, nil
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
