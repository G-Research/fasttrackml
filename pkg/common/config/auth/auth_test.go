package auth

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_NormalizeConfiguration(t *testing.T) {
	tests := []struct {
		name  string
		init  func() *Config
		check func(config *Config) bool
	}{
		{
			name: "TestAuthTypeUser",
			init: func() *Config {
				configPath := fmt.Sprintf("%s/configuration.yml", t.TempDir())
				// #nosec G304
				f, err := os.Create(configPath)
				assert.Nil(t, err)
				assert.Nil(t, f.Close())

				return &Config{
					AuthUsersConfig: configPath,
				}
			},
			check: func(config *Config) bool {
				return config.IsAuthTypeUser()
			},
		},
		{
			name: "TestAuthTypeOIDC",
			init: func() *Config {
				return &Config{
					AuthOIDCScopes:           []string{"scope1", "scope2"},
					AuthOIDCClientID:         "client_id",
					AuthOIDCClaimRoles:       "groups",
					AuthOIDCClientSecret:     "client_secret",
					AuthOIDCProviderEndpoint: "provider_endpoint",
				}
			},
			check: func(config *Config) bool {
				return config.IsAuthTypeOIDC()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.init()
			assert.Nil(t, config.NormalizeConfiguration())
			assert.True(t, tt.check(config))
		})
	}
}
