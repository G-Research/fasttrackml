package auth

// User represents object to store current user information.
type User struct {
	roles   []string
	isAdmin bool
}

// IsAdmin makes check that current user is Admin user.
func (u User) IsAdmin() bool {
	return u.isAdmin
}

// GetRoles returns current user roles.
func (u User) GetRoles() []string {
	return u.roles
}
