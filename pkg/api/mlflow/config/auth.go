package config

import "github.com/rotisserie/eris"

// supported list of authentication types.
const (
	AuthTypeOIDC  string = "oidc"
	AuthTypeRole  string = "role"
	AuthTypeBasic string = "basic"
)

type AuthConfig struct {
	AuthType     string
	AuthUsername string
	AuthPassword string
	AuthUserList []string
}

// IsAuthTypeOIDC makes check that current auth is AuthTypeOIDC.
func (c *AuthConfig) IsAuthTypeOIDC() bool {
	return c.AuthType == AuthTypeRole
}

// IsAuthTypeRole makes check that current auth is AuthTypeRole.
func (c *AuthConfig) IsAuthTypeRole() bool {
	return c.AuthType == AuthTypeRole
}

// IsAuthTypeBasic makes check that current auth is AuthTypeBasic.
func (c *AuthConfig) IsAuthTypeBasic() bool {
	return c.AuthType == AuthTypeBasic
}

// validateConfiguration validates service configuration for correctness.
func (c *AuthConfig) validateConfiguration() error {
	if c.AuthType != "" {
		if c.AuthType != AuthTypeOIDC && c.AuthType != AuthTypeRole && c.AuthType != AuthTypeBasic {
			return eris.Errorf(
				"provided auth type is incorrect. supported types are: %s, %s, %s",
				AuthTypeOIDC, AuthTypeRole, AuthTypeBasic,
			)
		}
	}
	return nil
}

// normalizeConfiguration normalizes auth configuration parameters.
func (c *AuthConfig) normalizeConfiguration() error {
	return nil
}
