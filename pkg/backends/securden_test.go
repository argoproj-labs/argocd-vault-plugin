package backends_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	securdensdk "github.com/SecurdenDevOps/securden-sdk"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/backends"
)

type mockSecurdenPasswordClient struct {
	password string
	body     string
	err      error
}

func (m *mockSecurdenPasswordClient) GetPassword(_ context.Context, _ int64, _ string, _ string) (string, string, error) {
	return m.password, m.body, m.err
}

func TestSecurdenGetSecretsInvalidAccountID(t *testing.T) {
	backend := backends.NewSecurdenBackend(&mockSecurdenPasswordClient{})

	_, err := backend.GetSecrets("not-an-account-id", "", nil)
	if err == nil {
		t.Fatal("expected invalid account id error")
	}

	expected := `invalid securden account_id "not-an-account-id"`
	if got := err.Error(); !strings.HasPrefix(got, expected) {
		t.Fatalf("expected error prefix %q, got %q", expected, got)
	}
}

func TestSecurdenGetSecretsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/secretsmanagement/get_password_via_tools" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.Header.Get("authtoken"); got != "test-token" {
			t.Fatalf("expected authtoken header, got %q", got)
		}
		if got := r.URL.Query().Get("account_id"); got != "2000000001800" {
			t.Fatalf("expected account_id query param, got %q", got)
		}
		if got := r.URL.Query().Get("reason"); got != "Argo CD secret rendering" {
			t.Fatalf("expected reason query param, got %q", got)
		}
		if got := r.URL.Query().Get("ticket_id"); got != "INC-123" {
			t.Fatalf("expected ticket_id query param, got %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"password":"super-secret"}`))
	}))
	defer server.Close()

	cfg := securdensdk.NewConfiguration(server.URL)
	cfg.SetAuthToken("test-token")

	backend := backends.NewSecurdenBackend(backends.NewSecurdenSDKClient(
		securdensdk.NewAPIClient(cfg),
		"Argo CD secret rendering",
		"INC-123",
	))

	data, err := backend.GetSecrets("2000000001800", "", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := data["password"]; got != "super-secret" {
		t.Fatalf("expected password to be returned, got %#v", got)
	}
}

func TestSecurdenGetIndividualSecretUnknownKey(t *testing.T) {
	backend := backends.NewSecurdenBackend(&mockSecurdenPasswordClient{password: "super-secret"})

	_, err := backend.GetIndividualSecret("2000000001800", "username", "", nil)
	if err == nil {
		t.Fatal("expected missing key error")
	}

	expected := `securden key "username" not found for account_id=2000000001800`
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

func TestSecurdenGetSecretsAPIFailure(t *testing.T) {
	backend := backends.NewSecurdenBackend(&mockSecurdenPasswordClient{
		body: `{"message":"upstream failure"}`,
		err:  fmt.Errorf("500 Internal Server Error"),
	})

	_, err := backend.GetSecrets("2000000001800", "", nil)
	if err == nil {
		t.Fatal("expected API failure error")
	}

	expected := `securden API failed for account_id=2000000001800: 500 Internal Server Error; response={"message":"upstream failure"}`
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}
