package config

import (
	"time"

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
func (c ServiceConfig) Validate() bool {
	return true
}
