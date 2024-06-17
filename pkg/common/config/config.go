package config

import (
	"net/url"
	"path"
	"path/filepath"
	"slices"
	"time"

	"github.com/rotisserie/eris"
	"github.com/spf13/viper"

	"github.com/G-Research/fasttrackml/pkg/common/config/auth"
)

// Config represents main service configuration.
type Config struct {
	Auth                  auth.Config
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
	RunLogOutputMax       int
	RunLogOutputRetain    time.Duration
}

// NewConfig creates a new instance of Config.
func NewConfig() *Config {
	return &Config{
		Auth: auth.Config{
			AuthUsername:             viper.GetString("auth-username"),
			AuthPassword:             viper.GetString("auth-password"),
			AuthUsersConfig:          viper.GetString("auth-users-config"),
			AuthOIDCScopes:           viper.GetStringSlice("auth-oidc-scopes"),
			AuthOIDCAdminRole:        viper.GetString("auth-oidc-admin-role"),
			AuthOIDCClientID:         viper.GetString("auth-oidc-client-id"),
			AuthOIDCClaimRoles:       viper.GetString("auth-oidc-claim-roles"),
			AuthOIDCClientSecret:     viper.GetString("auth-oidc-client-secret"),
			AuthOIDCProviderEndpoint: viper.GetString("auth-oidc-provider-endpoint"),
		},
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
		RunLogOutputMax:       viper.GetInt("log-output-max"),
		RunLogOutputRetain:    viper.GetDuration("log-output-retention"),
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

	if err := c.Auth.ValidateConfiguration(); err != nil {
		return eris.Wrap(err, "error validating auth configuration")
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

	if err := c.Auth.NormalizeConfiguration(); err != nil {
		return eris.Wrap(err, "error normalizing auth configuration")
	}

	return nil
}
