package auth

import (
	"github.com/spf13/viper"

	"github.com/G-Research/fasttrackml/pkg/common/auth/oidc"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
)

// supported list of authentication types.
const (
	TypeOIDC string = "oidc"
	TypeUser string = "user"
)

// Config represents Auth configuration.
type Config struct {
	AuthType                  string
	AuthUsername              string
	AuthPassword              string
	AuthOIDCClient            *oidc.Client
	AuthUsersConfig           string
	AuthOIDCClientID          string
	AuthOIDCClientSecret      string
	AuthOIDCScopes            []string
	AuthOIDCAdminRole         string
	AuthOIDCClaimRoles        string
	AuthOIDCProviderEndpoint  string
	AuthParsedUserPermissions *models.UserPermissions
}

// NewConfig creates a new instance of Config.
func NewConfig() *Config {
	return &Config{
		AuthUsername:             viper.GetString("auth-username"),
		AuthPassword:             viper.GetString("auth-password"),
		AuthUsersConfig:          viper.GetString("auth-users-config"),
		AuthOIDCScopes:           viper.GetStringSlice("auth-oidc-scopes"),
		AuthOIDCAdminRole:        viper.GetString("auth-oidc-admin-role"),
		AuthOIDCClientID:         viper.GetString("auth-oidc-client-id"),
		AuthOIDCClaimRoles:       viper.GetString("auth-oidc-claim-roles"),
		AuthOIDCClientSecret:     viper.GetString("auth-oidc-client-secret"),
		AuthOIDCProviderEndpoint: viper.GetString("auth-oidc-provider-endpoint"),
	}
}

// IsAuthTypeOIDC makes check that current auth is TypeOIDC.
func (c *Config) IsAuthTypeOIDC() bool {
	return c.AuthType == TypeOIDC
}

// IsAuthTypeUser makes check that current auth is TypeUser.
func (c *Config) IsAuthTypeUser() bool {
	return c.AuthType == TypeUser
}
