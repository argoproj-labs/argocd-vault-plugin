package backends

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
	"google.golang.org/grpc"
)

const keySecretTemplate = "%s#%s"

type mockYCLPayloadServiceClient struct {
	secrets map[string]*lockbox.Payload
}

func newMockYCLPayloadServiceClient() mockYCLPayloadServiceClient {
	return mockYCLPayloadServiceClient{
		secrets: map[string]*lockbox.Payload{},
	}
}

func (m mockYCLPayloadServiceClient) setSecret(secretID, key, version, value string) {
	keySecret := fmt.Sprintf(keySecretTemplate, secretID, version)
	payload, exists := m.secrets[keySecret]
	if !exists {
		payload = new(lockbox.Payload)
	}

	pe := &lockbox.Payload_Entry{}
	pe.SetKey(key)
	pe.SetTextValue(value)
	payload.SetEntries(append(payload.Entries, pe))

	m.secrets[keySecret] = payload
}

func (m mockYCLPayloadServiceClient) Get(_ context.Context, in *lockbox.GetPayloadRequest, _ ...grpc.CallOption) (*lockbox.Payload, error) {
	keySecret := fmt.Sprintf(keySecretTemplate, in.GetSecretId(), in.GetVersionId())
	payload, exists := m.secrets[keySecret]
	if !exists {
		return nil, fmt.Errorf("secret with id %q not found ", in.GetSecretId())
	}

	return payload, nil
}

func TestYandexCloudLockbox_GetIndividualSecret(t *testing.T) {
	client := newMockYCLPayloadServiceClient()
	client.setSecret("abc", "def", "", "1")
	client.setSecret("abc", "def", "2", "2")
	ycl := NewYandexCloudLockboxBackend(client)

	type args struct {
		secretID string
		key      string
		version  string
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "Get individual secret",
			args: args{
				secretID: "abc",
				key:      "def",
			},
			want:    "1",
			wantErr: false,
		},
		{
			name: "Get individual secret with version",
			args: args{
				secretID: "abc",
				key:      "def",
				version:  "2",
			},
			want:    "2",
			wantErr: false,
		},
		{
			name: "Get not exisiting individual secret",
			args: args{
				secretID: "abc",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ycl.GetIndividualSecret(tt.args.secretID, tt.args.key, tt.args.version, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetIndividualSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetIndividualSecret() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYandexCloudLockbox_GetSecrets(t *testing.T) {
	client := newMockYCLPayloadServiceClient()
	client.setSecret("abc", "def", "", "1")
	client.setSecret("abc", "def", "2", "2")
	ycl := NewYandexCloudLockboxBackend(client)

	type args struct {
		secretID string
		version  string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "Get secrets",
			args: args{
				secretID: "abc",
			},
			want:    map[string]interface{}{"def": "1"},
			wantErr: false,
		},
		{
			name: "Get secrets with version",
			args: args{
				secretID: "abc",
				version:  "2",
			},
			want:    map[string]interface{}{"def": "2"},
			wantErr: false,
		},
		{
			name: "Get not existing secrets",
			args: args{
				secretID: "abcde",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ycl.GetSecrets(tt.args.secretID, tt.args.version, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSecrets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSecrets() got = %v, want %v", got, tt.want)
			}
		})
	}
}
