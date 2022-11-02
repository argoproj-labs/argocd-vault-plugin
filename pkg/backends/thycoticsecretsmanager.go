package backends

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	thycoticsecretserver "github.com/thycotic/tss-sdk-go/server"
)

// ThycoticSecretServer is a struct for working with a Thycotic Secrets Manager backend
type ThycoticSecretServer struct {
	Client *thycoticsecretserver.Server
}

// NewThycoticSecretServerBackend initializes a new Thycotic Secrets Manager backend
func NewThycoticSecretServerBackend(client *thycoticsecretserver.Server) *ThycoticSecretServer {
	return &ThycoticSecretServer{
		Client: client,
	}
}

// Login does nothing as a "login" is handled on the instantiation of the Thycotic sdk
func (a *ThycoticSecretServer) Login() error {
	return nil
}

// GetSecrets gets secrets from Thycotic Secret Server and returns the formatted data
// Currently there is no implementation present for versions nor annotations
func (a *ThycoticSecretServer) GetSecrets(path string, version string, annotations map[string]string) (map[string]interface{}, error) {

	// Thycotic users pass the path of a secret
	// ex: <path:123#username>
	// So we query the secret id, and return a map

	input, err := strconv.Atoi(path)
	if err != nil {
		return nil, fmt.Errorf("could not read path %s", path)
	}
	secret, err := a.Client.Secret(input)
	if err != nil {
		return nil, fmt.Errorf("could not access secret %s", path)
	}
	utils.VerboseToStdErr("Thycotic Secret Server getting path %s", path)

	secret_json, err := json.MarshalIndent(secret, "", "  ")

	validation := make(map[string]interface{})

	if secret != nil {
		err := json.Unmarshal(secret_json, &validation)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("could not decode secret json %s", secret_json)
	}

	utils.VerboseToStdErr("Thycotic Secret Server decoding json %s", secret)

	secret_map := make(map[string]interface{})

	for index := range secret.Fields {
		secret_map[secret.Fields[index].FieldName] = secret.Fields[index].ItemValue
		secret_map[secret.Fields[index].Slug] = secret.Fields[index].ItemValue
	}

	utils.VerboseToStdErr("Thycotic Secret Server constructeed map %s", secret_map)
	return secret_map, nil

}

// GetIndividualSecret will get the specific secret (placeholder) from the SM backend
// For Thycotic Secret Server, we only support placeholders replaced from the k/v pairs of a secret which cannot be individually addressed
// So, we use GetSecrets and extract the specific placeholder we want
func (v *ThycoticSecretServer) GetIndividualSecret(kvpath, secret, version string, annotations map[string]string) (interface{}, error) {
	data, err := v.GetSecrets(kvpath, version, annotations)
	if err != nil {
		return nil, err
	}
	return data[secret], nil
}
