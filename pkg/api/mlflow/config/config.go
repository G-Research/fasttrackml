package config

import (
	"net/url"
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
func (c ServiceConfig) Validate() error {
	if err := c.validateArtifactRoot(); err != nil {
		return eris.Wrap(err, "error validating service configuration")
	}
	return nil
}

func (c ServiceConfig) validateArtifactRoot() error {
	parsedArtifactRoot, err := url.Parse(c.ArtifactRoot)
	if err != nil {
		return eris.Wrap(err, "error parsing `artifact-root` flag")
	}
	switch parsedArtifactRoot.Scheme {
	case "s3":
		if parsedArtifactRoot.User != nil || parsedArtifactRoot.RawQuery != "" || parsedArtifactRoot.RawFragment != "" {
			return eris.New("incorrect format of `artifact-root` flag. has to be s3://bucket_name")
		}
	}
	return nil
}
