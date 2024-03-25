package auth

// supported list of authentication types.
const (
	TypeOIDC string = "oidc"
	TypeUser string = "user"
)

type Config struct {
	AuthType        string
	AuthUsername    string
	AuthPassword    string
	AuthUsersConfig string
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
	}
	return nil
}
