package oidc

import "slices"

// User represents object to store current user information.
type User struct {
	Groups []string
}

// IsAdmin makes check that current user is Admin user.
func (u User) IsAdmin() bool {
	return slices.Contains(u.Groups, "admin")
}
