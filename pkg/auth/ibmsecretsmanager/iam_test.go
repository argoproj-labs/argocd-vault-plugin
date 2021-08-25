package ibmsecretsmanager_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/IBM/argocd-vault-plugin/pkg/auth/ibmsecretsmanager"
	"github.com/IBM/argocd-vault-plugin/pkg/helpers"
)

// MockClient is the mock client
type MockClient struct{}

// Do is the mock client's `Do` func
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	json := `{"access_token":"123"}`

	// create a new reader with that JSON
	r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
	return &http.Response{
		StatusCode: 200,
		Body:       r,
	}, nil
}

// Need to find a way to mock GitHub Auth within Vault
func TestIBMAuth(t *testing.T) {
	cluster := helpers.CreateTestAuthVault(t)
	defer cluster.Cleanup()

	c := &MockClient{}
	ibm := ibmsecretsmanager.NewIAMAuth("abc", c)

	err := ibm.Authenticate(cluster.Cores[0].Client)
	if err != nil {
		t.Fatalf("expected no errors but got: %s", err)
	}
}
