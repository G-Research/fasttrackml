package models

import "fmt"

// BasicAuthToken represents object to store auth information related to Basic Auth.
type BasicAuthToken struct {
	roles map[string]struct{}
}

// HasAdminAccess makes check that user has admin permissions to access to the requested resource.
func (p BasicAuthToken) HasAdminAccess() bool {
	if _, ok := p.roles["admin"]; ok {
		return true
	}
	return false
}

// HasUserAccess makes check that user has permission to access to the requested namespace.
func (p BasicAuthToken) HasUserAccess(namespace string) bool {
	if _, ok := p.roles[fmt.Sprintf("ns:%s", namespace)]; !ok {
		return ok
	}
	return true
}

// GetRoles returns User roles assigned to current Auth token.
func (p BasicAuthToken) GetRoles() (map[string]struct{}, bool) {
	return p.roles, true
}

// UserPermissions represents model to store user permissions data.
type UserPermissions struct {
	data map[string]map[string]struct{}
}

// NewUserPermissions creates new instance of UserPermissions object.
func NewUserPermissions(data map[string]map[string]struct{}) *UserPermissions {
	return &UserPermissions{
		data: data,
	}
}

// GetData returns current permissions data.
func (p UserPermissions) GetData() map[string]map[string]struct{} {
	return p.data
}

func (p UserPermissions) ValidateAuthToken(authToken string) (*BasicAuthToken, bool) {
	if authToken == "" {
		return nil, false
	}

	roles, ok := p.data[authToken]
	if !ok {
		return nil, ok
	}

	return &BasicAuthToken{
		roles: roles,
	}, true
}
