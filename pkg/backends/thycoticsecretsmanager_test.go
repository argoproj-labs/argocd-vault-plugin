package backends_test

import (
	"fmt"
	thycoticsecretserver "github.com/thycotic/tss-sdk-go/server"
	"github.com/IBM/argocd-vault-plugin/pkg/backends"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

type thycoticMockSender struct {
	DoFunc func(r *http.Request) (*http.Response, error)
}

func (m thycoticMockSender) Do(r *http.Request) (*http.Response, error) {
	return m.DoFunc(r)
}

func TestThycotic_GetSecrets(t *testing.T) {
	// secrets: list of key vault secrets (where foo is present and bar is absent)
	// foo: is a secret with a secret value
	// bar: is a secret with no secret value
	tt := map[string]struct {
		Body       string
		StatusCode int
	}{
		"secrets": {
			Body: `
					{
						"value": [
							{
								"contentType": "foobar",
								"id": "https://test.vault.azure.net/secrets/foo",
								"attributes": {
									"enabled": true,
									"created": 1629833926,
									"updated": 1629833926,
									"recoveryLevel": "Recoverable+Purgeable"
								},
								"tags": {}
							},
							{
								"id": "https://test.vault.azure.net/secrets/bar",
								"attributes": {
									"enabled": true,
									"created": 1629813653,
									"updated": 1629813653,
									"recoveryLevel": "Recoverable+Purgeable"
								},
								"tags": {
									"file-encoding": "utf-8"
								}
							}
						],
						"nextLink": null
					}`,
			StatusCode: 200,
		},
		"foo": {
			Body: `
					{
						"Name": "secret01",
						"FolderID": 1,
						"ID": 1,
						"SiteID": 1,
						"SecretTemplateID": 1,
						"SecretPolicyID": -1,
						"Active": true,
						"CheckedOut": false,
						"CheckOutEnabled": false,
						"Items": [
							{
								"ItemID": 1,
								"FieldID": 1,
								"FileAttachmentID": 0,
								"FieldDescription": "The secret field",
								"FieldName": "foo",
								"Filename": "",
								"ItemValue": "bar",
								"Slug": "field",
								"IsFile": false,
								"IsNotes": false,
								"IsPassword": false
							}
						]
					}`,
			StatusCode: 200,
		},
		"bar": {
			Body: `
					{
						"errorCode": "API_AccessDenied",
						"message": "Access Denied"
					}`,
			StatusCode: 400,
		},
	}

	// Setup client and mock Sender
	sender := &thycoticMockSender{}
	basicClientThycotic, _ := thycoticsecretserver.New(thycoticsecretserver.Configuration{
		Credentials: thycoticsecretserver.UserCredential{
			Username: `foo`,
			Password: `bar`,
		},
		ServerURL: `http://localhost/SecretServer`,
	})
	//basicClientThycotic.Sender = sender

	// DoFunc returns our mocked data when Do is called
	sender.DoFunc = func(r *http.Request) (*http.Response, error) {
		u, err := url.Parse(fmt.Sprintf("%s", r.URL))
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}
		if fmt.Sprintf("%s", u.Path) == "/secrets" {
			return &http.Response{
				StatusCode: tt["secrets"].StatusCode,
				Body:       io.NopCloser(strings.NewReader(tt["secrets"].Body)),
			}, nil
		} else if fmt.Sprintf("%s", u.Path) == "/secrets/bar/versions" {
			return &http.Response{
				StatusCode: tt["foobar"].StatusCode,
				Body:       io.NopCloser(strings.NewReader(tt["foobar"].Body)),
			}, nil
		} else if fmt.Sprintf("%s", u.Path) == "/secrets/bar/33740fc26214497f8904d93f20f7db6c" {
			return &http.Response{
				StatusCode: tt["bar_version"].StatusCode,
				Body:       io.NopCloser(strings.NewReader(tt["bar_version"].Body)),
			}, nil
		} else if fmt.Sprintf("%s", u.Path) == "/secrets/bar/33740fc26214497f8904d93f20f7db6b" {
			return &http.Response{
				StatusCode: tt["bar_disabled"].StatusCode,
				Body:       io.NopCloser(strings.NewReader(tt["bar_disabled"].Body)),
			}, nil
		} else {
			s := strings.Split(u.Path, "/")[2]
			return &http.Response{
				StatusCode: tt[s].StatusCode,
				Body:       io.NopCloser(strings.NewReader(tt[s].Body)),
			}, nil
		}
	}

	kv := backends.NewThycoticSecretServerBackend(basicClientThycotic)

	t.Run("Thycotic retrieve secrets no version", func(t *testing.T) {

		secretList, err := kv.GetSecrets("1", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"foo": "bar",
		}

		if !reflect.DeepEqual(expected, secretList) {
			t.Errorf("expected: %s, got: %s.", expected, secretList)
		}

	})

}
