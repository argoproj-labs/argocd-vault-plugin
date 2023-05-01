package backends

import (
	"encoding/json"
	"fmt"
	"strconv"

	delineasecretserver "github.com/DelineaXPM/tss-sdk-go/v2/server"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
)

// DelineaSecretServer is a struct for working with a Delinea Secrets Manager backend
type DelineaSecretServer struct {
	Client *delineasecretserver.Server
}

// NewDelineaSecretServerBackend initializes a new Delinea Secrets Manager backend
func NewDelineaSecretServerBackend(client *delineasecretserver.Server) *DelineaSecretServer {
	return &DelineaSecretServer{
		Client: client,
	}
}

// Login does nothing as a "login" is handled on the instantiation of the Delinea sdk
func (a *DelineaSecretServer) Login() error {
	return nil
}

// GetSecrets gets secrets from Delinea Secret Server and returns the formatted data
// Currently there is no implementation present for versions nor annotations
func (a *DelineaSecretServer) GetSecrets(path string, version string, annotations map[string]string) (map[string]interface{}, error) {

	// Delinea users pass the path of a secret
	// ex: <path:123#username>
	// So we query the secret id, and return a map

	input, err := strconv.Atoi(path)
	if err != nil {
		return nil, fmt.Errorf("could not read path %s, error: %s", path, err)
	}
	secret, err := a.Client.Secret(input)
	if err != nil {
		return nil, fmt.Errorf("could not access secret %s, error: %s", path, err)
	}
	utils.VerboseToStdErr("Delinea Secret Server getting path %s", path)

	secret_json, err := json.MarshalIndent(secret, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("enable to parse json secret %s, error: %s", path, err)
	}

	validation := make(map[string]interface{})

	if secret != nil {
		err := json.Unmarshal(secret_json, &validation)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("could not decode secret json %s", secret_json)
	}

	utils.VerboseToStdErr("Delinea Secret Server decoding json %s", secret)

	secret_map := make(map[string]interface{})

	for index := range secret.Fields {
		secret_map[secret.Fields[index].FieldName] = secret.Fields[index].ItemValue
		secret_map[secret.Fields[index].Slug] = secret.Fields[index].ItemValue
	}

	utils.VerboseToStdErr("Delinea Secret Server constructed map %s", secret_map)
	return secret_map, nil

}

// GetIndividualSecret will get the specific secret (placeholder) from the SM backend
// For Delinea Secret Server, we only support placeholders replaced from the k/v pairs of a secret which cannot be individually addressed
// So, we use GetSecrets and extract the specific placeholder we want
func (v *DelineaSecretServer) GetIndividualSecret(kvpath, secret, version string, annotations map[string]string) (interface{}, error) {
	data, err := v.GetSecrets(kvpath, version, annotations)
	if err != nil {
		return nil, err
	}
	return data[secret], nil
}
