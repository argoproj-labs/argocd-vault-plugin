package helpers

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func Factory(ctx context.Context, c *logical.BackendConfig) (logical.Backend, error) {
	b := Backend(c)
	if err := b.Setup(ctx, c); err != nil {
		return nil, err
	}
	return b, nil
}

type backend struct {
	*framework.Backend
}

func Backend(c *logical.BackendConfig) *backend {
	var b backend

	b.Backend = &framework.Backend{
		BackendType: logical.TypeCredential,
		AuthRenew:   b.pathAuthRenew,
		PathsSpecial: &logical.Paths{
			Unauthenticated: []string{"login"},
		},
		Paths: []*framework.Path{
			&framework.Path{
				Pattern: "login",
				Fields: map[string]*framework.FieldSchema{
					"token": &framework.FieldSchema{
						Type: framework.TypeString,
					},
				},
				Callbacks: map[logical.Operation]framework.OperationFunc{
					logical.UpdateOperation: b.pathAuthLogin,
				},
			},
		},
	}

	return &b
}

func (b *backend) pathAuthLogin(_ context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	return &logical.Response{
		Auth: &logical.Auth{
			InternalData: map[string]interface{}{
				"secret_value": "abcd1234",
			},
			LeaseOptions: logical.LeaseOptions{
				TTL:       30 * time.Hour,
				MaxTTL:    60 * time.Hour,
				Renewable: true,
			},
		},
	}, nil
}

func (b *backend) pathAuthRenew(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	if req.Auth == nil {
		return nil, errors.New("request auth was nil")
	}

	secretValue := req.Auth.InternalData["secret_value"].(string)
	if secretValue != "abcd1234" {
		return nil, errors.New("internal data does not match")
	}

	return framework.LeaseExtend(30*time.Hour, 60*time.Hour, b.System())(ctx, req, d)
}
