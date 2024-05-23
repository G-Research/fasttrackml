package config

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"slices"
	"time"

	"github.com/oauth2-proxy/mockoidc"
	"github.com/rotisserie/eris"
	"github.com/spf13/viper"

	"github.com/G-Research/fasttrackml/pkg/common/auth/oidc"
	"github.com/G-Research/fasttrackml/pkg/common/config/auth"
)

// Config represents main service configuration.
type Config struct {
	Auth                  *auth.Config
	DevMode               bool
	ListenAddress         string
	DefaultArtifactRoot   string
	S3EndpointURI         string
	GSEndpointURI         string
	DatabaseURI           string
	DatabaseReset         bool
	DatabasePoolMax       int
	DatabaseMigrate       bool
	DatabaseSlowThreshold time.Duration
	LiveUpdatesEnabled    bool
}

// NewConfig creates a new instance of Config.
func NewConfig() *Config {
	return &Config{
		Auth:                  auth.NewConfig(),
		DevMode:               viper.GetBool("dev-mode"),
		ListenAddress:         viper.GetString("listen-address"),
		DefaultArtifactRoot:   viper.GetString("default-artifact-root"),
		S3EndpointURI:         viper.GetString("s3-endpoint-uri"),
		GSEndpointURI:         viper.GetString("gs-endpoint-uri"),
		DatabaseURI:           viper.GetString("database-uri"),
		DatabaseReset:         viper.GetBool("database-reset"),
		DatabasePoolMax:       viper.GetInt("database-pool-max"),
		DatabaseMigrate:       viper.GetBool("database-migrate"),
		DatabaseSlowThreshold: viper.GetDuration("database-slow-threshold"),
		LiveUpdatesEnabled:    viper.GetBool("live-updates-enabled"),
	}
}

// Validate validates service configuration.
func (c *Config) Validate() error {
	if err := c.validateConfiguration(); err != nil {
		return eris.Wrap(err, "error validating service configuration")
	}
	if err := c.normalizeConfiguration(); err != nil {
		return eris.Wrap(err, "error normalizing service configuration")
	}
	return nil
}

// validateConfiguration validates service configuration for correctness.
func (c *Config) validateConfiguration() error {
	// 1. validate DefaultArtifactRoot configuration parameter for correctness and valid values.
	parsed, err := url.Parse(c.DefaultArtifactRoot)
	if err != nil {
		return eris.Wrap(err, "error parsing 'default-artifact-root' flag")
	}

	if parsed.User != nil || parsed.RawQuery != "" || parsed.RawFragment != "" {
		return eris.New("incorrect format of 'default-artifact-root' flag")
	}

	if !slices.Contains([]string{"", "file", "s3", "gs"}, parsed.Scheme) {
		return eris.New("unsupported schema of 'default-artifact-root' flag")
	}

	return nil
}

// normalizeConfiguration normalizes service configuration parameters.
func (c *Config) normalizeConfiguration() error {
	parsed, err := url.Parse(c.DefaultArtifactRoot)
	if err != nil {
		return eris.Wrap(err, "error parsing 'default-artifact-root' flag")
	}
	switch parsed.Scheme {
	case "", "file":
		absoluteArtifactRoot, err := filepath.Abs(path.Join(parsed.Host, parsed.Path))
		if err != nil {
			return eris.Wrapf(err, "error getting absolute path for 'default-artifact-root': %s", c.DefaultArtifactRoot)
		}
		c.DefaultArtifactRoot = "file://" + absoluteArtifactRoot
	}

	// create temporary OIDC mock server here and initialize application configuration.
	// this is a temporary solution just for testing.
	oidcMockServer, err := mockoidc.Run()
	if err != nil {
		return eris.Wrap(err, "error creating oidc mock server")
	}
	oidcMockServer.QueueUser(&mockoidc.MockUser{
		Email:  "test.user@example.com",
		Groups: []string{"group1", "group2"},
	})
	c.Auth.AuthOIDCScopes = []string{"openid", "groups"}
	c.Auth.AuthOIDCClientID = oidcMockServer.ClientID
	c.Auth.AuthOIDCAdminRole = "admin"
	c.Auth.AuthOIDCClaimRoles = "groups"
	c.Auth.AuthOIDCClientSecret = oidcMockServer.ClientSecret
	c.Auth.AuthOIDCProviderEndpoint = oidcMockServer.Addr() + mockoidc.IssuerBase

	switch {
	case c.Auth.AuthUsersConfig != "":
		c.Auth.AuthType = auth.TypeUser
		parsedUserPermissions, err := auth.Load(c.Auth.AuthUsersConfig)
		if err != nil {
			return eris.Wrapf(err, "error loading auth user configuration from file: %s", c.Auth.AuthUsersConfig)
		}
		c.Auth.AuthParsedUserPermissions = parsedUserPermissions
	case c.Auth.AuthOIDCClientID != "" && c.Auth.AuthOIDCClientSecret != "" && c.Auth.AuthOIDCProviderEndpoint != "":
		oidcClient, err := oidc.NewClient(
			context.Background(),
			fmt.Sprintf("http://%s", c.ListenAddress),
			c.Auth.AuthOIDCProviderEndpoint, c.Auth.AuthOIDCClientID, c.Auth.AuthOIDCClientSecret,
			c.Auth.AuthOIDCClaimRoles, c.Auth.AuthOIDCAdminRole,
			c.Auth.AuthOIDCScopes,
		)
		if err != nil {
			return eris.Wrap(err, "error creating OIDC client")
		}
		c.Auth.AuthType = auth.TypeOIDC
		c.Auth.AuthOIDCClient = oidcClient
	}
	return nil
}
