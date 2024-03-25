package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_NormalizeConfiguration(t *testing.T) {
	tests := []struct {
		name       string
		config     *Config
		configType string
	}{
		{
			name: "TestAuthTypeUser",
			config: &Config{
				AuthUsersConfig: "/path/to/file",
			},
			configType: TypeUser,
		},
		{
			name: "TestAuthTypeBasic",
			config: &Config{
				AuthUsername: "username",
				AuthPassword: "password",
			},
			configType: TypeBasic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Nil(t, tt.config.NormalizeConfiguration())
			assert.Equal(t, tt.configType, tt.config.AuthType)
		})
	}
}
