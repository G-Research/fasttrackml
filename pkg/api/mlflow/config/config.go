package config

import (
	"net/url"
	"path"
	"path/filepath"
	"time"

	"golang.org/x/exp/slices"

	"github.com/rotisserie/eris"
	"github.com/spf13/viper"
)

// ServiceConfig represents main service configuration.
type ServiceConfig struct {
	DevMode               bool
	ListenAddress         string
	AuthUsername          string
	AuthPassword          string
	ArtifactRoot          string
	S3EndpointURI         string
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
		ArtifactRoot:          viper.GetString("artifact-root"),
		S3EndpointURI:         viper.GetString("s3-endpoint-uri"),
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
	if err := c.normaliseConfiguration(); err != nil {
		return eris.Wrap(err, "error normalising service configuration")
	}
	return nil
}

// validateConfiguration validates service configuration for correctness.
func (c *ServiceConfig) validateConfiguration() error {
	// 1. validate ArtifactRoot configuration parameter for correctness and valid values.
	parsed, err := url.Parse(c.ArtifactRoot)
	if err != nil {
		return eris.Wrap(err, `error parsing 'artifact-root' flag`)
	}

	if parsed.User != nil || parsed.RawQuery != "" || parsed.RawFragment != "" {
		return eris.New(`incorrect format of 'artifact-root' flag`)
	}

	if !slices.Contains([]string{"", "file", "s3"}, parsed.Scheme) {
		return eris.New(`unsupported schema of 'artifact-root' flag`)
	}

	return nil
}

// normaliseConfiguration normalizes service configuration parameters.
func (c *ServiceConfig) normaliseConfiguration() error {
	parsed, err := url.Parse(c.ArtifactRoot)
	if err != nil {
		return eris.Wrap(err, `error parsing 'artifact-root' flag`)
	}
	switch parsed.Scheme {
	case "s3":
		return nil
	case "", "file":
		absoluteArtifactRoot, err := filepath.Abs(path.Join(parsed.Host, parsed.Path))
		if err != nil {
			return eris.Wrapf(err, `error getting absolute path for 'artifact-root': %s`, c.ArtifactRoot)
		}
		c.ArtifactRoot = absoluteArtifactRoot
	}
	return nil
}
