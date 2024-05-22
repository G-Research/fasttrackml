package oidc

import (
	"context"
	"fmt"
	"slices"

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
	adminRole    string
	claimRoles   string
	verifier     *oidc.IDTokenVerifier
	oauth2Config *oauth2.Config
}

// NewClient creates a new OIDC client.
func NewClient(
	ctx context.Context,
	listenAddress string,
	providerEndpoint, clientID, clientSecret string,
	claimRoles, adminRole string,
	scopes []string,
) (*Client, error) {
	provider, err := oidc.NewProvider(ctx, providerEndpoint)
	if err != nil {
		return nil, eris.Wrap(err, "error creating OIDC provider")
	}
	return &Client{
		adminRole:  adminRole,
		claimRoles: claimRoles,
		verifier: provider.Verifier(
			&oidc.Config{
				ClientID:        clientID,
				SkipIssuerCheck: true,
			},
		),
		oauth2Config: &oauth2.Config{
			Scopes:       scopes,
			Endpoint:     provider.Endpoint(),
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  fmt.Sprintf("%s/callback/oidc", listenAddress),
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

	data, ok := claims[c.claimRoles]
	if !ok {
		return nil, eris.Errorf("claim property: %s not found", c.claimRoles)
	}

	roles, err := ConvertAndNormaliseRoles(data)
	if err != nil {
		return nil, eris.Wrapf(err, "error converting claim %s property", c.claimRoles)
	}
	return &User{
		roles:   roles,
		isAdmin: slices.Contains(roles, c.adminRole),
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
