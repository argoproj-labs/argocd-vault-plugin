package backends

import (
	"context"
	"fmt"
	"strings"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

const (
	AWS_PS_CURRENT  string = "AWSCURRENT"
	AWS_PS_PREVIOUS string = "AWSPREVIOUS"
)

type AWSSSMParameterStoreIface interface {
	GetParametersByPath(ctx context.Context,
		params *ssm.GetParametersByPathInput,
		optFns ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error)

	GetParameter(ctx context.Context,
		params *ssm.GetParameterInput,
		optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

// TODO Comment
type AWSSSMParameterStore struct {
	Client AWSSSMParameterStoreIface
}

// NewAWSSSMParameterStoreBackend initializes a new AWS System Manager Parameter Store backend
func NewAWSSSMParameterStoreBackend(client AWSSSMParameterStoreIface) *AWSSSMParameterStore {
	return &AWSSSMParameterStore{
		Client: client,
	}
}

// Login does nothing as a "login" is handled on the instantiation of the aws sdk
func (a *AWSSSMParameterStore) Login() error {
	return nil
}

// GetSecrets gets secrets from aws ssm parameter store and returns the formatted data
func (a *AWSSSMParameterStore) GetSecrets(path, version string, annotations map[string]string) (map[string]interface{}, error) {
	input := &ssm.GetParametersByPathInput{
		Path:           aws.String(path),
		Recursive:      aws.Bool(false),
		WithDecryption: aws.Bool(true),
	}

	utils.VerboseToStdErr("AWS SSM Parameter Store getting secrets by path %s", path)
	result, err := a.Client.GetParametersByPath(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	utils.VerboseToStdErr("AWS SSM Parameter Store get secret response %v", &result)

	data := make(map[string]interface{})

	if result.Parameters != nil {
		for _, parameter := range result.Parameters {
			// extract the parameter name from the path
			split := strings.Split(*parameter.Name, "/")
			parameterName := split[len(split)-1]

			data[parameterName] = *parameter.Value
		}
	} else {
		return nil, fmt.Errorf("Could not find secret by path %s", path)
	}

	return data, nil
}

// GetSecrets gets one specific secret from aws ssm parameter store and returns the formatted data
func (a *AWSSSMParameterStore) GetIndividualSecret(kvpath, secret, version string, annotations map[string]string) (interface{}, error) {

	fullName := fmt.Sprintf("%v/%v", kvpath, secret)
	input := &ssm.GetParameterInput{
		Name:           aws.String(fullName),
		WithDecryption: aws.Bool(true),
	}

	if version != "" {
		*input.Name = fmt.Sprintf("%v:%v", *input.Name, version)
	}

	utils.VerboseToStdErr("AWS SSM Parameter Store getting secret %s", kvpath)
	result, err := a.Client.GetParameter(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	utils.VerboseToStdErr("AWS SSM Parameter Store get secret response %v", &result)

	data := make(map[string]interface{})

	if result.Parameter.Value != nil {
		data["parameter"] = *result.Parameter.Value
	} else {
		return nil, fmt.Errorf("Could not find secret %s", kvpath)
	}

	return data["parameter"], nil
}
