package oidc

import (
	"context"
	"slices"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/common/config/auth"
)

// ClientProvider provides an interface to work with OIDC provider.
type ClientProvider interface {
	// Verify makes Access Token verification.
	Verify(ctx context.Context, accessToken string) (*User, error)
}

// Client represents OIDC client.
type Client struct {
	config   *auth.Config
	verifier *oidc.IDTokenVerifier
}

// NewClient creates new OIDC client
func NewClient(ctx context.Context, config *auth.Config) (*Client, error) {
	provider, err := oidc.NewProvider(ctx, config.AuthOIDCProviderEndpoint)
	if err != nil {
		return nil, eris.Wrap(err, "error creating OIDC provider")
	}

	return &Client{
		config:   config,
		verifier: provider.Verifier(&oidc.Config{ClientID: config.AuthOIDCClientID, SkipIssuerCheck: true}),
	}, nil
}

// Verify makes Access Token verification.
func (c Client) Verify(ctx context.Context, accessToken string) (*User, error) {
	idToken, err := c.verifier.Verify(ctx, accessToken)
	if err != nil {
		return nil, eris.Wrap(err, "error verifying access token")
	}
	// Extract custom claims.
	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		return nil, eris.Wrap(err, "error extracting token claims")
	}

	data, ok := claims[c.config.AuthOIDCClaimRoles]
	if !ok {
		return nil, eris.Errorf("claim property: %s not found", c.config.AuthOIDCClaimRoles)
	}

	roles, err := ConvertAndNormaliseRoles(data)
	if err != nil {
		return nil, eris.Wrapf(err, "error converting claim %s property", c.config.AuthOIDCClaimRoles)
	}
	return &User{
		roles:   roles,
		isAdmin: slices.Contains(roles, c.config.AuthOIDCAdminRole),
	}, nil
}
