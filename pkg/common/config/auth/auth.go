package auth

import (
	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
)

// supported list of authentication types.
const (
	TypeOIDC string = "oidc"
	TypeUser string = "user"
)

type Config struct {
	AuthUsername              string
	AuthPassword              string
	AuthUsersConfig           string
	AuthOIDCClientID          string
	AuthOIDCClientSecret      string
	AuthOIDCScopes            []string
	AuthOIDCAdminRole         string
	AuthOIDCClaimRoles        string
	AuthOIDCProviderEndpoint  string
	AuthParsedUserPermissions *models.UserPermissions
}

// IsAuthTypeOIDC makes check that current auth is TypeOIDC.
func (c *Config) IsAuthTypeOIDC() bool {
	return c.AuthOIDCClientID != "" &&
		len(c.AuthOIDCScopes) > 0 &&
		c.AuthOIDCClaimRoles != "" &&
		c.AuthOIDCClientSecret != "" &&
		c.AuthOIDCProviderEndpoint != ""
}

// IsAuthTypeUser makes check that current auth is TypeUser.
func (c *Config) IsAuthTypeUser() bool {
	return c.AuthParsedUserPermissions != nil
}

// ValidateConfiguration validates service configuration for correctness.
func (c *Config) ValidateConfiguration() error {
	return nil
}

// NormalizeConfiguration normalizes auth configuration parameters.
func (c *Config) NormalizeConfiguration() error {
	if c.AuthUsersConfig != "" {
		parsedUserPermissions, err := Load(c.AuthUsersConfig)
		if err != nil {
			return eris.Wrapf(err, "error loading auth user configuration from file: %s", c.AuthUsersConfig)
		}
		c.AuthParsedUserPermissions = parsedUserPermissions
	}
	return nil
}
