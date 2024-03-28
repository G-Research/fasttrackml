package auth

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_NormalizeConfiguration(t *testing.T) {
	tests := []struct {
		name       string
		init       func() *Config
		configType string
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
			configType: TypeUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.init()
			assert.Nil(t, config.NormalizeConfiguration())
			assert.Equal(t, tt.configType, config.AuthType)
		})
	}
}
