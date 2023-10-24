package config

import (
	"net/url"
	"path"
	"path/filepath"
	"slices"
	"time"

	"github.com/rotisserie/eris"
	"github.com/spf13/viper"
)

// ServiceConfig represents main service configuration.
type ServiceConfig struct {
	DevMode               bool
	ListenAddress         string
	AuthUsername          string
	AuthPassword          string
	DefaultArtifactRoot   string
	S3EndpointURI         string
	GSEndpointURI         string
	DatabaseURI           string
	DatabaseReset         bool
	DatabasePoolMax       int
	DatabaseMigrate       bool
	DatabaseSlowThreshold time.Duration
}

// NewServiceConfig creates new instance of ServiceConfig.
func NewServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		DevMode:               viper.GetBool("dev-mode"),
		ListenAddress:         viper.GetString("listen-address"),
		AuthUsername:          viper.GetString("auth-username"),
		AuthPassword:          viper.GetString("auth-password"),
		DefaultArtifactRoot:   viper.GetString("default-artifact-root"),
		S3EndpointURI:         viper.GetString("s3-endpoint-uri"),
		GSEndpointURI:         viper.GetString("gs-endpoint-uri"),
		DatabaseURI:           viper.GetString("database-uri"),
		DatabaseReset:         viper.GetBool("database-reset"),
		DatabasePoolMax:       viper.GetInt("database-pool-max"),
		DatabaseMigrate:       viper.GetBool("database-migrate"),
		DatabaseSlowThreshold: viper.GetDuration("database-slow-threshold"),
	}
}

// Validate validates service configuration.
func (c *ServiceConfig) Validate() error {
	if err := c.validateConfiguration(); err != nil {
		return eris.Wrap(err, "error validating service configuration")
	}
	if err := c.normalizeConfiguration(); err != nil {
		return eris.Wrap(err, "error normalizing service configuration")
	}
	return nil
}

// validateConfiguration validates service configuration for correctness.
func (c *ServiceConfig) validateConfiguration() error {
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
func (c *ServiceConfig) normalizeConfiguration() error {
	parsed, err := url.Parse(c.DefaultArtifactRoot)
	if err != nil {
		return eris.Wrap(err, "error parsing 'default-artifact-root' flag")
	}
	switch parsed.Scheme {
	case "s3", "gs":
		return nil
	case "", "file":
		absoluteArtifactRoot, err := filepath.Abs(path.Join(parsed.Host, parsed.Path))
		if err != nil {
			return eris.Wrapf(err, "error getting absolute path for 'default-artifact-root': %s", c.DefaultArtifactRoot)
		}
		c.DefaultArtifactRoot = absoluteArtifactRoot
	}
	return nil
}
