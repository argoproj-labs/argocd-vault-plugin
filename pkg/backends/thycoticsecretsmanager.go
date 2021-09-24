package backends

import (
	"encoding/json"
	"fmt"
	"strconv"
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

	input, err := strconv.Atoi(path)
	secret, err := a.Client.Secret(input)
	if err != nil {
		return nil, fmt.Errorf("Could not access secret %s", path)
	}

	secret_json, err := json.MarshalIndent(secret,"", "  ")

	validation := make(map[string]interface{})

	if secret != nil {
		err := json.Unmarshal(secret_json, &validation)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("Could not decode secret json %s", path)
	}

	secret_map := make(map[string]interface{})

	for index := range secret.Fields {
		secret_map[secret.Fields[index].FieldName] = secret.Fields[index].ItemValue
	}

	return secret_map, nil

}
