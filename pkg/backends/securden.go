package backends

import (
	"context"
	"fmt"
	"strconv"

	securdensdk "github.com/SecurdenDevOps/securden-sdk"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
)

type SecurdenPasswordClient interface {
	GetPassword(ctx context.Context, accountID int64, reason, ticketID string) (string, string, error)
}

// Securden is a struct for working with the Securden backend.
type Securden struct {
	Client SecurdenPasswordClient
}

type securdenSDKClient struct {
	client          *securdensdk.APIClient
	defaultReason   string
	defaultTicketID string
}

// NewSecurdenBackend initializes a new Securden backend.
func NewSecurdenBackend(client SecurdenPasswordClient) *Securden {
	return &Securden{
		Client: client,
	}
}

// NewSecurdenSDKClient wraps the Securden SDK client behind the backend interface.
func NewSecurdenSDKClient(client *securdensdk.APIClient, defaultReason, defaultTicketID string) SecurdenPasswordClient {
	return &securdenSDKClient{
		client:          client,
		defaultReason:   defaultReason,
		defaultTicketID: defaultTicketID,
	}
}

// Login does nothing because the SDK uses the configured auth token directly.
func (s *Securden) Login() error {
	return nil
}

// GetSecrets fetches the account password and exposes it as a password-only map.
// TODO: Extend this if Securden exposes additional stable fields needed by AVP.
func (s *Securden) GetSecrets(path string, version string, annotations map[string]string) (map[string]interface{}, error) {
	accountID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid securden account_id %q: %w", path, err)
	}

	utils.VerboseToStdErr("Securden getting password for account_id=%d", accountID)

	var reason string
	var ticketID string
	if annotations != nil {
		reason = annotations["AVP_SECURDEN_REASON"]
		ticketID = annotations["AVP_SECURDEN_TICKET_ID"]
	}

	password, responseBody, err := s.Client.GetPassword(context.Background(), accountID, reason, ticketID)
	if err != nil {
		errMsg := fmt.Sprintf("securden API failed for account_id=%d: %v", accountID, err)
		if responseBody != "" {
			errMsg = fmt.Sprintf("%s; response=%s", errMsg, responseBody)
		}
		return nil, fmt.Errorf("%s", errMsg)
	}

	return map[string]interface{}{
		"password": password,
	}, nil
}

// GetIndividualSecret returns the requested field from the Securden response map.
func (s *Securden) GetIndividualSecret(path, secret, version string, annotations map[string]string) (interface{}, error) {
	data, err := s.GetSecrets(path, version, annotations)
	if err != nil {
		return nil, err
	}

	value, ok := data[secret]
	if !ok {
		return nil, fmt.Errorf("securden key %q not found for account_id=%s", secret, path)
	}

	return value, nil
}

func (s *securdenSDKClient) GetPassword(ctx context.Context, accountID int64, reason, ticketID string) (string, string, error) {
	req := s.client.DefaultAPI.GetPassword(ctx).AccountId(accountID)
	if reason == "" {
		reason = s.defaultReason
	}
	if ticketID == "" {
		ticketID = s.defaultTicketID
	}
	if reason != "" {
		req = req.Reason(reason)
	}
	if ticketID != "" {
		req = req.TicketId(ticketID)
	}

	resp, body, err := req.Execute()
	if err != nil {
		return "", body, err
	}

	return resp.GetPassword(), body, nil
}
