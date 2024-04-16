package auth

import (
	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/common/db/models"
)

// supported list of authentication types.
const (
	TypeOIDC string = "oidc"
	TypeUser string = "user"
)

type Config struct {
	AuthType                  string
	AuthUsername              string
	AuthPassword              string
	AuthUsersConfig           string
	AuthOIDCClientID          string
	AuthOIDCClientSecret      string
	AuthOIDCProviderEndpoint  string
	AuthParsedUserPermissions *models.UserPermissions
}

// IsAuthTypeOIDC makes check that current auth is TypeOIDC.
func (c *Config) IsAuthTypeOIDC() bool {
	return c.AuthType == TypeUser
}

// IsAuthTypeUser makes check that current auth is TypeUser.
func (c *Config) IsAuthTypeUser() bool {
	return c.AuthType == TypeUser
}

// ValidateConfiguration validates service configuration for correctness.
func (c *Config) ValidateConfiguration() error {
	return nil
}

// NormalizeConfiguration normalizes auth configuration parameters.
func (c *Config) NormalizeConfiguration() error {
	switch {
	case c.AuthUsersConfig != "":
		c.AuthType = TypeUser
		parsedUserPermissions, err := Load(c.AuthUsersConfig)
		if err != nil {
			return eris.Wrapf(err, "error loading auth user configuration from file: %s", c.AuthUsersConfig)
		}
		c.AuthParsedUserPermissions = parsedUserPermissions
	case c.AuthOIDCClientID != "" && c.AuthOIDCClientSecret != "" && c.AuthOIDCProviderEndpoint != "":
		c.AuthType = TypeOIDC
	}
	return nil
}
