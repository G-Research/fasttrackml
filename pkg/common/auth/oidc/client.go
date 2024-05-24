package oidc

import (
	"context"
	"fmt"
	"slices"

	"github.com/G-Research/fasttrackml/pkg/common/config"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rotisserie/eris"
	"golang.org/x/oauth2"
)

// ClientProvider provides an interface to work with OIDC provider.
type ClientProvider interface {
	// Verify makes Access Token verification.
	Verify(ctx context.Context, accessToken string) (*User, error)
	// Exchange converts an authorization code into a token.
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
	// GetOauth2Config returns oauth2 configuration.
	GetOauth2Config() *oauth2.Config
}

// Client represents OIDC client.
type Client struct {
	config       *config.Config
	verifier     *oidc.IDTokenVerifier
	oauth2Config *oauth2.Config
}

// NewClient creates a new OIDC client.
func NewClient(ctx context.Context, config *config.Config,
) (*Client, error) {
	provider, err := oidc.NewProvider(ctx, config.Auth.AuthOIDCProviderEndpoint)
	if err != nil {
		return nil, eris.Wrap(err, "error creating OIDC provider")
	}
	return &Client{
		config: config,
		verifier: provider.Verifier(
			&oidc.Config{
				ClientID:        config.Auth.AuthOIDCClientID,
				SkipIssuerCheck: true,
			},
		),
		oauth2Config: &oauth2.Config{
			Scopes:       config.Auth.AuthOIDCScopes,
			Endpoint:     provider.Endpoint(),
			ClientID:     config.Auth.AuthOIDCClientID,
			ClientSecret: config.Auth.AuthOIDCClientSecret,
			RedirectURL:  fmt.Sprintf("%s/callback/oidc", NormaliseListenAddress(config.ListenAddress)),
		},
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

	data, ok := claims[c.config.Auth.AuthOIDCClaimRoles]
	if !ok {
		return nil, eris.Errorf("claim property: %s not found", c.config.Auth.AuthOIDCClaimRoles)
	}

	roles, err := ConvertAndNormaliseRoles(data)
	if err != nil {
		return nil, eris.Wrapf(err, "error converting claim %s property", c.config.Auth.AuthOIDCClaimRoles)
	}
	return &User{
		roles:   roles,
		isAdmin: slices.Contains(roles, c.config.Auth.AuthOIDCAdminRole),
	}, nil
}

// Exchange converts an authorization code into a token.
func (c Client) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	oauth2Token, err := c.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, eris.Wrap(err, "error converting an authorization code into a token")
	}
	return oauth2Token, nil
}

// GetOauth2Config returns oauth2 configuration.
func (c Client) GetOauth2Config() *oauth2.Config {
	return c.oauth2Config
}
