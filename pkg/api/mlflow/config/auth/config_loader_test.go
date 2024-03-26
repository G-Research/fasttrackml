package auth

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/G-Research/fasttrackml/pkg/common/db/models"
)

func TestLoad_Ok(t *testing.T) {
	tests := []struct {
		name        string
		init        func() string
		permissions *models.UserPermissions
	}{
		{
			name: "TestLoadConfigurationWithYmlExtension",
			init: func() string {
				cfg := YamlConfig{
					Users: []YamlUserConfig{
						{
							Name: "user1",
							Roles: []string{
								"ns:namespace1",
								"ns:namespace2",
							},
							Password: "user1password",
						},
					},
				}
				data, err := yaml.Marshal(cfg)
				assert.Nil(t, err)

				configPath := fmt.Sprintf("%s/configuration.yml", t.TempDir())
				// #nosec G304
				f, err := os.Create(configPath)
				assert.Nil(t, err)
				_, err = f.Write(data)
				assert.Nil(t, err)
				assert.Nil(t, f.Close())
				return configPath
			},
			permissions: models.NewUserPermissions(map[string]map[string]struct{}{
				"dXNlcjE6dXNlcjFwYXNzd29yZA==": {
					"ns:namespace1": struct{}{},
					"ns:namespace2": struct{}{},
				},
			}),
		},
		{
			name: "TestLoadConfigurationWithYamlExtension",
			init: func() string {
				cfg := YamlConfig{
					Users: []YamlUserConfig{
						{
							Name: "user2",
							Roles: []string{
								"ns:namespace3",
								"ns:namespace4",
							},
							Password: "user2password",
						},
					},
				}
				data, err := yaml.Marshal(cfg)
				assert.Nil(t, err)

				configPath := fmt.Sprintf("%s/configuration.yaml", t.TempDir())
				// #nosec G304
				f, err := os.Create(configPath)
				assert.Nil(t, err)
				_, err = f.Write(data)
				assert.Nil(t, err)
				assert.Nil(t, f.Close())
				return configPath
			},
			permissions: models.NewUserPermissions(map[string]map[string]struct{}{
				"dXNlcjI6dXNlcjJwYXNzd29yZA==": {
					"ns:namespace3": struct{}{},
					"ns:namespace4": struct{}{},
				},
			}),
		},
		{
			name: "TestLoadConfigurationUserPasswordExistsInENV",
			init: func() string {
				cfg := YamlConfig{
					Users: []YamlUserConfig{
						{
							Name: "user3",
							Roles: []string{
								"ns:namespace3",
								"ns:namespace4",
							},
							Password: "${LOAD_FROM_ENV}",
						},
					},
				}
				data, err := yaml.Marshal(cfg)
				assert.Nil(t, err)

				configPath := fmt.Sprintf("%s/configuration.yaml", t.TempDir())
				// #nosec G304
				f, err := os.Create(configPath)
				assert.Nil(t, err)
				_, err = f.Write(data)
				assert.Nil(t, err)
				assert.Nil(t, f.Close())

				assert.Nil(t, os.Setenv("LOAD_FROM_ENV", "user2password"))
				return configPath
			},
			permissions: models.NewUserPermissions(map[string]map[string]struct{}{
				"dXNlcjM6dXNlcjJwYXNzd29yZA==": {
					"ns:namespace3": struct{}{},
					"ns:namespace4": struct{}{},
				},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := tt.init()
			userPermissions, err := Load(configPath)
			assert.Nil(t, err)
			assert.Equal(t, tt.permissions.GetData(), userPermissions.GetData())
		})
	}
}

func TestLoad_Error(t *testing.T) {
	configPath := fmt.Sprintf("%s/configuration.unsupported-extension", t.TempDir())
	// #nosec G304
	_, err := os.Create(configPath)
	assert.Nil(t, err)

	_, err = Load(configPath)
	assert.Equal(t, "unsupported user configuration file type", err.Error())
}

func TestUserPermissions_HasAccess_Ok(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		namespace   string
		permissions *models.UserPermissions
	}{
		{
			name:      "TestUserPermissionsUserHasPermissions",
			token:     "token",
			namespace: "namespace1",
			permissions: models.NewUserPermissions(map[string]map[string]struct{}{
				"token": {
					"ns:namespace1": struct{}{},
				},
			}),
		},
		{
			name:      "TestUserPermissionsUserHasAdminRole",
			token:     "token",
			namespace: "namespace1",
			permissions: models.NewUserPermissions(map[string]map[string]struct{}{
				"token": {
					"admin": struct{}{},
				},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authToken, isValid := tt.permissions.ValidateAuthToken(tt.token)
			assert.True(t, isValid)
			assert.True(t, authToken.HasUserAccess(tt.namespace))
		})
	}
}

func TestUserPermissions_HasAccess_Error(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		namespace   string
		permissions *models.UserPermissions
	}{
		{
			name:        "TestUserPermissionsAuthTokenIsEmpty",
			token:       "",
			permissions: models.NewUserPermissions(map[string]map[string]struct{}{}),
		},
		{
			name:  "TestUserPermissionsAuthTokenNotFound",
			token: "not-existing-token",
			permissions: models.NewUserPermissions(map[string]map[string]struct{}{
				"token": {},
			}),
		},
		{
			name:      "TestUserPermissionsUserHasNoAccessToNamespace",
			token:     "token",
			namespace: "namespace1",
			permissions: models.NewUserPermissions(map[string]map[string]struct{}{
				"token": {
					"ns:namespace2": struct{}{},
				},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authToken, isValid := tt.permissions.ValidateAuthToken(tt.token)
			assert.False(t, isValid)
			assert.False(t, authToken.HasUserAccess(tt.namespace))
		})
	}
}
