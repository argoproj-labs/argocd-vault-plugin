package backends_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/backends"
	ksm "github.com/keeper-security/secrets-manager-go/core"
)

// mockResults is an internal data structure used by the MockKeeperClient.
type mockResults struct {
	Records   []*ksm.Record
	ExpectErr error
}

// MockKeeperClient is a mock backends.KeeperClient
type MockKeeperClient struct {
	mocks map[string]mockResults
}

// GetSecrets is the fake function that returns a secret based on the
// pre-programmed mocks field on the MockKeeperClient struct.
func (c MockKeeperClient) GetSecrets(ids []string) ([]*ksm.Record, error) {
	if len(ids) != 1 {
		return nil, fmt.Errorf("invalid ids given")
	}

	path := ids[0]

	response, ok := c.mocks[path]
	if !ok {
		return nil, fmt.Errorf("no response prepared for the given path: %s", path)
	}

	if response.ExpectErr != nil {
		return nil, response.ExpectErr
	}

	return response.Records, nil
}

func recordFromJSON(data string) *ksm.Record {
	return &ksm.Record{
		RecordDict: ksm.JsonToDict(data),
	}
}

func TestKeeperSecretsManager_GetSecrets(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "should handle a secret of type login",
			args: args{
				data: `{
	"uid": "some-uid",
	"title": "some-title",
	"type": "login",
	"fields": [
		{
			"label": "login",
			"type": "login",
			"value": [
				"some-secret-username"
			]
		},
		{
			"label": "password",
			"type": "password",
			"value": [
				"some-secret-password"
			]
		},
		{
			"label": "url",
			"type": "url",
			"value": []
		},
		{
			"label": "fileRef",
			"type": "fileRef",
			"value": []
		},
		{
			"label": "oneTimeCode",
			"type": "oneTimeCode",
			"value": []
		}
	],
	"custom": [],
	"files": []
}`,
			},
			want: map[string]interface{}{
				"login":    "some-secret-username",
				"password": "some-secret-password",
			},
		},
		{
			name: "should handle a secret of type databaseCredentials",
			args: args{
				data: `
{
    "uid": "some-uid",
    "title": "some-title",
    "type": "databaseCredentials",
    "fields": [
        {
            "label": "type",
            "type": "text",
            "value": [
                "some-database-type"
            ]
        },
        {
            "label": "host",
            "type": "host",
            "value": [
                {
                    "hostName": "some-hostname",
                    "port": "some-port"
                }
            ]
        },
        {
            "label": "login",
            "type": "login",
            "value": [
                "some-login"
            ]
        },
        {
            "label": "password",
            "type": "password",
            "value": [
                "some-password"
            ]
        },
        {
            "label": "fileRef",
            "type": "fileRef",
            "value": []
        }
    ],
    "custom": [],
    "files": []
}`,
			},
			want: map[string]interface{}{
				"host": map[string]interface{}{
					"hostName": "some-hostname",
					"port":     "some-port",
				},
				"login":    "some-login",
				"password": "some-password",
				"type":     "some-database-type",
			},
		},
		{
			name: "should handle a secret of type encryptedNotes",
			args: args{
				data: `
{
    "uid": "some-uid",
    "title": "some-title",
    "type": "encryptedNotes",
    "fields": [
        {
            "label": "note",
            "type": "note",
            "value": [
                "some-value"
            ]
        },
        {
            "label": "date",
            "type": "date",
            "value": []
        },
        {
            "label": "fileRef",
            "type": "fileRef",
            "value": []
        }
    ],
    "custom": [],
    "files": []
}`,
			},
			want: map[string]interface{}{
				"note": "some-value",
			},
		},
		{
			name: "should handle a secret with custom fields",
			args: args{
				data: `
{
    "uid": "some-uid",
    "title": "some-title",
    "type": "encryptedNotes",
    "fields": [
        {
            "label": "note",
            "type": "note",
            "value": [
                "some-value"
            ]
        },
        {
            "label": "date",
            "type": "date",
            "value": []
        },
        {
            "label": "fileRef",
            "type": "fileRef",
            "value": []
        }
    ],
    "custom": [
		{
			"label":"custom",
			"type":"text",
			"value":[
				"some-custom-secret"
			]
		}
	],
    "files": []
}`,
			},
			want: map[string]interface{}{
				"note":   "some-value",
				"custom": "some-custom-secret",
			},
		},
		{
			name: "should not overwrite a built in field when a custom field of the same label exists",
			args: args{
				data: `
{
    "uid": "some-uid",
    "title": "some-title",
    "type": "encryptedNotes",
    "fields": [
        {
            "label": "note",
            "type": "note",
            "value": [
                "some-value"
            ]
        },
        {
            "label": "date",
            "type": "date",
            "value": []
        },
        {
            "label": "fileRef",
            "type": "fileRef",
            "value": []
        }
    ],
    "custom": [
		{
			"label":"note",
			"type":"text",
			"value":[
				"some-custom-value"
			]
		}
	],
    "files": []
}`,
			},
			want: map[string]interface{}{
				"note": "some-value",
			},
		},
		{
			name: "should omit fields that have multiple values",
			args: args{
				data: `
{
    "uid": "some-uid",
    "title": "some-title",
    "type": "some-type",
    "fields": [
        {
            "label": "note",
            "type": "note",
            "value": [
                "some-value"
            ]
        },
		{
			"label": "other-note",
            "type": "note",
            "value": [
                "some-value",
				"some-value2"
            ]
		},
        {
            "label": "date",
            "type": "date",
            "value": []
        },
        {
            "label": "fileRef",
            "type": "fileRef",
            "value": []
        }
    ],
    "custom": [],
    "files": []
}`,
			},
			want: map[string]interface{}{
				"note": "some-value",
			},
		},

		{
			name: "should omit fields that don't have a value",
			args: args{
				data: `
{
    "uid": "some-uid",
    "title": "some-title",
    "type": "some-type",
    "fields": [
        {
            "label": "note",
            "type": "note",
            "value": [
                "some-value"
            ]
        },
		{
			"label": "other-note",
            "type": "note"
		},
        {
            "label": "date",
            "type": "date",
            "value": []
        },
        {
            "label": "fileRef",
            "type": "fileRef",
            "value": []
        }
    ],
    "custom": [],
    "files": []
}`,
			},
			want: map[string]interface{}{
				"note": "some-value",
			},
		},
		{
			name: "should omit fields that don't have a type",
			args: args{
				data: `
{
    "uid": "some-uid",
    "title": "some-title",
    "type": "some-type",
    "fields": [
        {
            "label": "note",
            "type": "note",
            "value": [
                "some-value"
            ]
        },
		{
			"label": "other-note",
            "value": [
				"foo-bar"
			]
		},
        {
            "label": "date",
            "type": "date",
            "value": []
        },
        {
            "label": "fileRef",
            "type": "fileRef",
            "value": []
        }
    ],
    "custom": [],
    "files": []
}`,
			},
			want: map[string]interface{}{
				"note": "some-value",
			},
		},
		{
			name: "should omit fields that don't have a label or type",
			args: args{
				data: `
{
    "uid": "some-uid",
    "title": "some-title",
    "type": "some-type",
    "fields": [
        {
            "label": "note",
            "type": "note",
            "value": [
                "some-value"
            ]
        },
		{
            "value": [
				"foo-bar"
			]
		},
        {
            "label": "date",
            "type": "date",
            "value": []
        },
        {
            "label": "fileRef",
            "type": "fileRef",
            "value": []
        }
    ],
    "custom": [],
    "files": []
}`,
			},
			want: map[string]interface{}{
				"note": "some-value",
			},
		},
		{
			name: "should omit fields that don't have a value that is not a slice",
			args: args{
				data: `
{
    "uid": "some-uid",
    "title": "some-title",
    "type": "some-type",
    "fields": [
        {
            "label": "note",
            "type": "note",
            "value": [
                "some-value"
            ]
        },
		{
			"label": "other-note",
			"type": "text",
            "value": "foo-bar"
		},
        {
            "label": "date",
            "type": "date",
            "value": []
        },
        {
            "label": "fileRef",
            "type": "fileRef",
            "value": []
        }
    ],
    "custom": [],
    "files": []
}`,
			},
			want: map[string]interface{}{
				"note": "some-value",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := backends.NewKeeperSecretsManagerBackend(
				MockKeeperClient{
					mocks: map[string]mockResults{
						"path": {
							Records: []*ksm.Record{
								recordFromJSON(tt.args.data),
							},
						},
					},
				},
			)
			got, err := a.GetSecrets("path", "", nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("KeeperSecretsManager.GetSecrets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeeperSecretsManager.GetSecrets() = %v, want %v", got, tt.want)
			}
		})
	}
}
