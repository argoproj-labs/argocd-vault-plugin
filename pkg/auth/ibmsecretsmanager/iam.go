package ibmsecretsmanager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/IBM/argocd-vault-plugin/pkg/types"
	"github.com/IBM/argocd-vault-plugin/pkg/utils"
	"github.com/hashicorp/vault/api"
)

// IAMAuth is a struct for working with SecretManager that uses IAM
type IAMAuth struct {
	APIKey string
	Client types.HTTPClient
}

// NewIAMAuth initializes a new IAMAuth with api key
func NewIAMAuth(apikey string, client types.HTTPClient) *IAMAuth {
	iamAuth := &IAMAuth{
		APIKey: apikey,
		Client: client,
	}

	return iamAuth
}

// Authenticate authenticates with Vault using App Role and returns a token
func (i *IAMAuth) Authenticate(vaultClient *api.Client) error {
	accessToken, err := getAccessToken(i.APIKey, i.Client)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"token": accessToken,
	}

	data, err := vaultClient.Logical().Write("auth/ibmcloud/login", payload)
	if err != nil {
		return err
	}

	// If we cannot write the Vault token, we'll just have to login next time. Nothing showstopping.
	err = utils.SetToken(vaultClient, data.Auth.ClientToken)
	if err != nil {
		print(err)
	}

	return nil
}

func getAccessToken(apikey string, client types.HTTPClient) (string, error) {
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

	// Perform http request
	res, err := client.Do(req)
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
