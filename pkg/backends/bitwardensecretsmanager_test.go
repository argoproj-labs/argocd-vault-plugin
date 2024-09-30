package backends_test

import (
	"testing"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/backends"
	bitwarden "github.com/bitwarden/sdk-go"
)

type mockBitwardenSecretesClient struct{}

func (bw mockBitwardenSecretesClient) List(organizationID string) (*bitwarden.SecretIdentifiersResponse, error) {
	switch organizationID {
	case "58293c58-5666-11ef-91a2-67fcd9d549c7":
		return &bitwarden.SecretIdentifiersResponse{
			Data: []bitwarden.SecretIdentifierResponse{
				{
					ID:             "ce398fa2-5665-11ef-8916-97605d6da25b",
					Key:            "Human Readable Key",
					OrganizationID: organizationID,
				},
				{
					ID:             "98b6c8ee-5666-11ef-ac37-8742ac5fc78f",
					Key:            "Other Key",
					OrganizationID: organizationID,
				},
			},
		}, nil
	default:
		return nil, nil
	}
}

func (bw mockBitwardenSecretesClient) Get(secretID string) (*bitwarden.SecretResponse, error) {
	switch secretID {
	case "ce398fa2-5665-11ef-8916-97605d6da25b":
		projectID := "ddb13dae-5665-11ef-8583-f73233caa8df"
		return &bitwarden.SecretResponse{
			CreationDate:   "2022-11-17T15:55:18.005669100Z",
			ID:             secretID,
			Key:            "Human Readable Key",
			Note:           "",
			OrganizationID: "d4105690-5665-11ef-a058-c713a9374bb0",
			ProjectID:      &projectID,
			RevisionDate:   "2022-11-17T15:55:18.005669100Z",
			Value:          "my secret",
		}, nil
	case "98b6c8ee-5666-11ef-ac37-8742ac5fc78f":
		projectID := "ddb13dae-5665-11ef-8583-f73233caa8df"
		return &bitwarden.SecretResponse{
			CreationDate:   "2019-05-11T15:55:18.005669100Z",
			ID:             secretID,
			Key:            "Other Key",
			Note:           "",
			OrganizationID: "d4105690-5665-11ef-a058-c713a9374bb0",
			ProjectID:      &projectID,
			RevisionDate:   "2019-05-11T15:55:18.005669100Z",
			Value:          "my other secret",
		}, nil
	default:
		return nil, nil
	}

}

func TestBitwardenSecretsManager(t *testing.T) {
	sm := backends.NewBitwardenSecretsClient(mockBitwardenSecretesClient{})

	t.Run("Test Login", func(t *testing.T) {
		err := sm.Login()
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}
	})

	t.Run("GetIndividualSecret", func(t *testing.T) {
		secret, err := sm.GetIndividualSecret("58293c58-5666-11ef-91a2-67fcd9d549c7", "ce398fa2-5665-11ef-8916-97605d6da25b", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}
		if secret != "my secret" {
			t.Fatalf("expected secret value 'my secret' but got: %s", secret)
		}
	})

	t.Run("GetSecrets", func(t *testing.T) {
		secrets, err := sm.GetSecrets("58293c58-5666-11ef-91a2-67fcd9d549c7", "", map[string]string{})
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}
		if secrets["ce398fa2-5665-11ef-8916-97605d6da25b"] != "my secret" {
			t.Fatalf("expected 'my secret' but got: %s", secrets["ce398fa2-5665-11ef-8916-97605d6da25b"])
		}
		if secrets["98b6c8ee-5666-11ef-ac37-8742ac5fc78f"] != "my other secret" {
			t.Fatalf("expected 'my other secret' but got: %s", secrets["98b6c8ee-5666-11ef-ac37-8742ac5fc78f"])
		}
	})
}
