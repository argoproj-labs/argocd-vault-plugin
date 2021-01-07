package vault

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SecretManager is a struct for working with IBM Secret Manager
type SecretManager struct {
	IBMCloudAPIKey string
	*Client
}

// Login authenticates with IBM Cloud Secret Manager using IAM and returns a token
func (s *SecretManager) Login() error {
	accessToken, err := getAccessToken(s.IBMCloudAPIKey)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"token": accessToken,
	}

	data, err := s.Client.Write("auth/ibmcloud/login", payload)
	if err != nil {
		return err
	}

	SetToken(s.Client, data.Auth.ClientToken)
	return nil
}

// GetSecrets gets secrets from IBM Secret Manager and returns the formatted data
func (s *SecretManager) GetSecrets(path string) (map[string]interface{}, error) {
	data, err := s.Client.Read(path)
	if err != nil {
		return nil, err
	}

	// Make sure the secret exists
	if _, ok := data["secrets"]; !ok {
		return nil, fmt.Errorf("Could not find secrets at path %s", path)
	}

	// Get list of secrets
	secretList := data["secrets"].([]interface{})
	v := make([]string, 0, len(secretList))
	// Loop through secrets and get id
	// as getting the list of secrets does not include the payload
	for _, value := range secretList {
		secret := value.(map[string]interface{})
		if t, found := secret["id"]; found {
			v = append(v, t.(string))
		}
	}

	// Read each secret and get payload
	secrets := make(map[string]interface{})
	for _, j := range v {
		data, err := s.Client.Read(fmt.Sprintf("%s/%s", path, j))
		if err != nil {
			return nil, err
		}
		// Get name and data of secret and append to secrets map
		secretName := data["name"].(string)
		secretData := data["secret_data"].(map[string]interface{})
		secrets[secretName] = secretData["payload"]
	}

	return secrets, nil
}

func getAccessToken(apikey string) (string, error) {
	// Set url values to be added to the request
	urlValues := url.Values{}
	urlValues.Set("grant_type", "urn:ibm:params:oauth:grant-type:apikey")
	urlValues.Add("apikey", apikey)

	// Creating request to get access token
	req, err := http.NewRequest("POST", "https://iam.cloud.ibm.com/identity/token", strings.NewReader(urlValues.Encode()))
	if err != nil {
		fmt.Print(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	var httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	// Perform http request
	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	data := map[string]interface{}{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", err
	}

	return data["access_token"].(string), nil
}
