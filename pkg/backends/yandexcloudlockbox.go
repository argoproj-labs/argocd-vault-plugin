package backends

import (
	"context"
	"fmt"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
)

// YandexCloudLockbox is a struct for working with a Yandex Cloud lockbox backend
type YandexCloudLockbox struct {
	client lockbox.PayloadServiceClient
}

// NewYandexCloudLockboxBackend initializes a new Yandex Cloud lockbox backend
func NewYandexCloudLockboxBackend(client lockbox.PayloadServiceClient) *YandexCloudLockbox {
	return &YandexCloudLockbox{
		client: client,
	}
}

// Login does nothing as a "login" is handled on the instantiation of the lockbox
func (ycl *YandexCloudLockbox) Login() error {
	return nil
}

// GetSecrets gets secrets from lockbox and returns the formatted data
func (ycl *YandexCloudLockbox) GetSecrets(secretID string, version string, _ map[string]string) (map[string]interface{}, error) {
	req := &lockbox.GetPayloadRequest{
		SecretId: secretID,
	}

	if version != "" {
		req.SetVersionId(version)
	}

	utils.VerboseToStdErr("Yandex Cloud Lockbox getting secret %s at version %s", secretID, version)
	resp, err := ycl.client.Get(context.Background(), req)
	if err != nil {
		return nil, err
	}

	utils.VerboseToStdErr("Yandex Cloud Lockbox get secret response %v", resp)

	result := make(map[string]interface{}, len(resp.GetEntries()))
	for _, v := range resp.GetEntries() {
		result[v.GetKey()] = v.GetTextValue()
	}

	return result, nil
}

// GetIndividualSecret will get the specific secret (placeholder) from the lockbox backend
func (ycl *YandexCloudLockbox) GetIndividualSecret(secretID, key, version string, _ map[string]string) (interface{}, error) {
	secrets, err := ycl.GetSecrets(secretID, version, nil)
	if err != nil {
		return nil, err
	}

	secret, found := secrets[key]
	if !found {
		return nil, fmt.Errorf("secretID: %s, key: %s, version: %s not found", secretID, key, version)
	}

	return secret, nil
}
