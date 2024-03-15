package auth

// supported list of authentication types.
const (
	TypeOIDC  string = "oidc"
	TypeRBAC  string = "rbac"
	TypeBasic string = "basic"
)

type Config struct {
	AuthType           string
	AuthUsername       string
	AuthPassword       string
	AuthRBACConfigFile string
}

// IsAuthTypeOIDC makes check that current auth is TypeOIDC.
func (c *Config) IsAuthTypeOIDC() bool {
	return c.AuthType == TypeRBAC
}

// IsAuthTypeRBAC makes check that current auth is TypeRBAC.
func (c *Config) IsAuthTypeRBAC() bool {
	return c.AuthType == TypeRBAC
}

// IsAuthTypeBasic makes check that current auth is TypeBasic.
func (c *Config) IsAuthTypeBasic() bool {
	return c.AuthType == TypeBasic
}

// ValidateConfiguration validates service configuration for correctness.
func (c *Config) ValidateConfiguration() error {
	return nil
}

// NormalizeConfiguration normalizes auth configuration parameters.
func (c *Config) NormalizeConfiguration() error {
	switch {
	case c.AuthUsername != "" && c.AuthPassword != "":
		c.AuthType = TypeBasic
	case c.AuthRBACConfigFile != "":
		c.AuthType = TypeRBAC

	}
	return nil
}
