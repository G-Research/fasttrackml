package oidc

// User represents object to store current user information.
type User struct {
	roles   []string
	isAdmin bool
}

// Roles returns current user roles.
func (u User) Roles() []string {
	return u.roles
}

// IsAdmin makes check that current user is Admin user.
func (u User) IsAdmin() bool {
	return u.isAdmin
}
