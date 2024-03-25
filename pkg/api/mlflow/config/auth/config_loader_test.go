package auth

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestLoad_Ok(t *testing.T) {
	tests := []struct {
		name        string
		init        func() string
		permissions map[string]map[string]struct{}
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
			permissions: map[string]map[string]struct{}{
				"dXNlcjE6dXNlcjFwYXNzd29yZA==": {
					"ns:namespace1": struct{}{},
					"ns:namespace2": struct{}{},
				},
			},
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
			permissions: map[string]map[string]struct{}{
				"dXNlcjI6dXNlcjJwYXNzd29yZA==": {
					"ns:namespace3": struct{}{},
					"ns:namespace4": struct{}{},
				},
			},
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
			permissions: map[string]map[string]struct{}{
				"dXNlcjM6dXNlcjJwYXNzd29yZA==": {
					"ns:namespace3": struct{}{},
					"ns:namespace4": struct{}{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := tt.init()
			userPermissions, err := Load(configPath)
			assert.Nil(t, err)
			assert.Equal(t, tt.permissions, userPermissions.data)
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
		permissions *UserPermissions
	}{
		{
			name:      "TestUserPermissionsUserHasPermissions",
			token:     "token",
			namespace: "namespace1",
			permissions: &UserPermissions{data: map[string]map[string]struct{}{
				"token": {
					"ns:namespace1": struct{}{},
				},
			}},
		},
		{
			name:      "TestUserPermissionsUserHasAdminRole",
			token:     "token",
			namespace: "namespace1",
			permissions: &UserPermissions{data: map[string]map[string]struct{}{
				"token": {
					"admin": struct{}{},
				},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.permissions.HasUserAccess(tt.namespace, tt.token))
		})
	}
}

func TestUserPermissions_HasAccess_Error(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		namespace   string
		permissions *UserPermissions
	}{
		{
			name:        "TestUserPermissionsAuthTokenIsEmpty",
			token:       "",
			permissions: &UserPermissions{data: map[string]map[string]struct{}{}},
		},
		{
			name:  "TestUserPermissionsAuthTokenNotFound",
			token: "not-existing-token",
			permissions: &UserPermissions{data: map[string]map[string]struct{}{
				"token": {},
			}},
		},
		{
			name:      "TestUserPermissionsUserHasNoAccessToNamespace",
			token:     "token",
			namespace: "namespace1",
			permissions: &UserPermissions{data: map[string]map[string]struct{}{
				"token": {
					"ns:namespace2": struct{}{},
				},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.False(t, tt.permissions.HasUserAccess(tt.namespace, tt.token))
		})
	}
}