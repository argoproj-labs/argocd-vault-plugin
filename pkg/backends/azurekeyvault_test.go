package backends_test

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/keyvault/keyvault"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/backends"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

type mockSender struct {
	DoFunc func(r *http.Request) (*http.Response, error)
}

func (m mockSender) Do(r *http.Request) (*http.Response, error) {
	return m.DoFunc(r)
}

func TestAzureKeyVault_GetSecrets(t *testing.T) {
	// secrets: list of key vault secrets (where foo and bar is present)
	// foo: is a secret with a secret value
	// bar: is a secret with a secret value
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
						"value": "bar",
						"contentType": "foobar",
						"id": "https://test.vault.azure.net.test/secrets/foo/8f8da2e06c8240808ee439ff093803b5",
						"attributes": {
							"enabled": true,
							"created": 1629833926,
							"updated": 1629833926,
							"recoveryLevel": "Recoverable+Purgeable"
						},
						"tags": {}
					}`,
			StatusCode: 200,
		},
		"bar": {
			Body: `
					{
						"value": "baz",
						"id": "https://test.vault.azure.net.test/secrets/bar/33740fc26214497f8904d93f20f7db6d",
						"attributes": {
							"enabled": true,
							"created": 1629813653,
							"updated": 1629813653,
							"recoveryLevel": "Recoverable+Purgeable"
						},
						"tags": {
							"file-encoding": "utf-8"
						}
					}`,
			StatusCode: 200,
		},
		"bar_version": {
			Body: `
					{
						"value": "baz-version",
						"id": "https://test.vault.azure.net.test/secrets/bar/33740fc26214497f8904d93f20f7db6c",
						"attributes": {
							"enabled": true,
							"created": 1629813653,
							"updated": 1629813653,
							"recoveryLevel": "Recoverable+Purgeable"
						},
						"tags": {
							"file-encoding": "utf-8"
						}
					}`,
			StatusCode: 200,
		},
		"bar_disabled": {
			Body: `
					{
						"value": "baz-disabled",
						"id": "https://test.vault.azure.net.test/secrets/bar/33740fc26214497f8904d93f20f7db6b",
						"attributes": {
							"enabled": false,
							"created": 1629813653,
							"updated": 1629813653,
							"recoveryLevel": "Recoverable+Purgeable"
						},
						"tags": {
							"file-encoding": "utf-8"
						}
					}`,
			StatusCode: 200,
		},
		"foobar": {
			Body: `
					{
						"value": [
							{
								"value": "bar",
								"id": "https://test.vault.azure.net.test/secrets/bar/33740fc26214497f8904d93f20f7db6d",
								"attributes": {
									"enabled": true,
									"created": 1629813653,
									"updated": 1629813653,
									"recoveryLevel": "Recoverable+Purgeable"
								},
								"tags": {
									"file-encoding": "utf-8"
								}
							},
							{
								"value": "bar",
								"id": "https://test.vault.azure.net.test/secrets/bar/33740fc26214497f8904d93f20f7db6c",
								"attributes": {
									"enabled": true,
									"created": 1629813653,
									"updated": 1629813653,
									"recoveryLevel": "Recoverable+Purgeable"
								},
								"tags": {
									"file-encoding": "utf-8"
								}
							},
							{
								"value": "bar",
								"id": "https://test.vault.azure.net.test/secrets/bar/33740fc26214497f8904d93f20f7db6b",
								"attributes": {
									"enabled": false,
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
	}

	// Setup client and mock Sender
	sender := &mockSender{}
	basicClient := keyvault.New()
	basicClient.Sender = sender

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

	kv := backends.NewAzureKeyVaultBackend(basicClient)

	t.Run("Azure retrieve secrets no version", func(t *testing.T) {

		secretList, err := kv.GetSecrets("test", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"foo": "bar",
			"bar": "baz",
		}

		if !reflect.DeepEqual(expected, secretList) {
			t.Errorf("expected: %s, got: %s.", expected, secretList)
		}

	})

	t.Run("Azure retrieve secrets with version", func(t *testing.T) {

		// test version
		secretList, err := kv.GetSecrets("test", "33740fc26214497f8904d93f20f7db6c", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"bar": "baz-version",
		}

		if !reflect.DeepEqual(expected, secretList) {
			t.Errorf("expected: %s, got: %s.", expected, secretList)
		}

	})

	t.Run("Azure GetIndividualSecret", func(t *testing.T) {
		secret, err := kv.GetIndividualSecret("test", "bar", "33740fc26214497f8904d93f20f7db6c", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := "baz-version"

		if !reflect.DeepEqual(expected, secret) {
			t.Errorf("expected: %s, got: %s.", expected, secret)
		}
	})

	t.Run("Azure retrieve secrets with version disabled", func(t *testing.T) {

		// test disabled secret
		secretList, err := kv.GetSecrets("test", "33740fc26214497f8904d93f20f7db6b", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{}

		if !reflect.DeepEqual(expected, secretList) {
			t.Errorf("expected: %s, got: %s.", expected, secretList)
		}

	})
}
