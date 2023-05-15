package backends

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

const (
	AWS_PS_CURRENT  string = "AWSCURRENT"
	AWS_PS_PREVIOUS string = "AWSPREVIOUS"
)

type AWSSSMParameterStoreIface interface {
	GetParameter(ctx context.Context,
		params *ssm.GetParameterInput,
		optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

// TODO Comment
type AWSSSMParameterStore struct {
	Client AWSSSMParameterStoreIface
}

// NewAWSSSMParameterStoreBackend initializes a new AWS Secrets Manager backend
func NewAWSSSMParameterStoreBackend(client AWSSSMParameterStoreIface) *AWSSSMParameterStore {
	return &AWSSSMParameterStore{
		Client: client,
	}
}

// Login does nothing as a "login" is handled on the instantiation of the aws sdk
func (a *AWSSSMParameterStore) Login() error {
	return nil
}

// GetSecrets gets secrets from aws secrets manager and returns the formatted data
func (a *AWSSSMParameterStore) GetParameters(path, version string, annotations map[string]string) (map[string]interface{}, error) {
	input := &ssm.GetParameterInput{
		Name:           aws.String(path),
		WithDecryption: bool(true),
	}

	if version != "" {
		*input.Name = fmt.Sprintf("%v:%v", *input.Name, version)
	}

	utils.VerboseToStdErr("AWS SSM Parameter Store getting secret %s", path)
	result, err := a.Client.GetParameter(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	utils.VerboseToStdErr("AWS SSM Parameter Store get secret response %v", result)

	var dat map[string]interface{}

	if result.Parameter.Value != nil {
		err := json.Unmarshal([]byte(*result.Parameter.Value), &dat)
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
func (a *AWSSSMParameterStore) GetIndividualSecret(kvpath, secret, version string, annotations map[string]string) (interface{}, error) {
	data, err := a.GetParameters(kvpath, version, annotations)
	if err != nil {
		return nil, err
	}
	return data[secret], nil
}
