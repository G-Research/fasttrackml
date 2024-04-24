package auth

import (
	"context"
	"slices"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/common/config/auth"
)

// OIDCClientProvider provides an interface to work with OIDC provider.
type OIDCClientProvider interface {
	// Verify makes Access Token verification.
	Verify(ctx context.Context, accessToken string) (*User, error)
}

// OIDCClient represents OIDC client.
type OIDCClient struct {
	config   *auth.Config
	verifier *oidc.IDTokenVerifier
}

// NewOIDCClient creates new OIDC client
func NewOIDCClient(ctx context.Context, config *auth.Config) (*OIDCClient, error) {
	provider, err := oidc.NewProvider(ctx, config.AuthOIDCProviderEndpoint)
	if err != nil {
		return nil, eris.Wrap(err, "error creating OIDC provider")
	}

	return &OIDCClient{
		config:   config,
		verifier: provider.Verifier(&oidc.Config{ClientID: config.AuthOIDCClientID, SkipIssuerCheck: true}),
	}, nil
}

// Verify makes Access Token verification.
func (c OIDCClient) Verify(ctx context.Context, accessToken string) (*User, error) {
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
