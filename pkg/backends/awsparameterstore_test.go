package backends_test

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"reflect"
	"strings"
	"testing"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/backends"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type mockSSMParameterStoreClient struct {
	backends.AWSSSMParameterStoreIface
}

func (m *mockSSMParameterStoreClient) GetParameter(ctx context.Context, input *ssm.GetParameterInput, options ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	data := &ssm.GetParameterOutput{
		Parameter: &types.Parameter{},
	}

	// first entry in split is the name, seccond the version
	split := strings.Split(*input.Name, ":")
	switch split[0] {
	case "test/test-secret":
		// verify if version is set
		if len(split) < 2 {
			string := "{\"test-secret\":\"current-value\"}"
			data.Parameter.Value = &string
		} else {
			string := "{\"test-secret\":\"previous-value\"}"
			data.Parameter.Value = &string
		}
	}

	return data, nil
}

func (m *mockSSMParameterStoreClient) GetParametersByPath(ctx context.Context, input *ssm.GetParametersByPathInput, options ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error) {

	data := &ssm.GetParametersByPathOutput{
		Parameters: []types.Parameter{},
	}

	switch *input.Path {
	case "test":
		parameters := []types.Parameter{
			{
				Name:  aws.String("test-secret"),
				Value: aws.String("current-value"),
				Type:  types.ParameterTypeSecureString,
			},
		}

		data = &ssm.GetParametersByPathOutput{
			Parameters: parameters,
		}
	}

	return data, nil
}

func TestAWSSSMParameterStoreGetSecrets(t *testing.T) {
	ps := backends.NewAWSSSMParameterStoreBackend(&mockSSMParameterStoreClient{})

	t.Run("Get secrets", func(t *testing.T) {
		data, err := ps.GetSecrets("test", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"test-secret": "current-value",
		}

		if !reflect.DeepEqual(expected, data) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}
	})

	t.Run("AWS GetIndividualSecret", func(t *testing.T) {
		secret, err := ps.GetIndividualSecret("test", "test-secret", "previous", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := "{\"test-secret\":\"previous-value\"}"

		if !reflect.DeepEqual(expected, secret) {
			t.Errorf("expected: %s, got: %s.", expected, secret)
		}
	})

	t.Run("Get secrets at specific version", func(t *testing.T) {
		data, err := ps.GetIndividualSecret("test", "test-secret", "123", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := "{\"test-secret\":\"previous-value\"}"

		if !reflect.DeepEqual(expected, data) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}
	})
}

func TestAWSSSMParameterStoreEmptyIfNoSecret(t *testing.T) {
	sm := backends.NewAWSSSMParameterStoreBackend(&mockSSMParameterStoreClient{})

	_, err := sm.GetSecrets("empty", "", map[string]string{})
	if err == nil {
		t.Fatalf("expected an error but got nil")
	}

	if err.Error() != "Could not find secret by path empty" {
		t.Errorf("expected error: %s, got: %s.", "Could not find secret empty", err.Error())
	}
}
