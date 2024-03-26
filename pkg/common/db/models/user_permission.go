package models

import "fmt"

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

// HasAccess makes check that user has permission to access to the requested namespace.
func (p UserPermissions) HasAccess(namespace string, authToken string) bool {
	if authToken == "" {
		return false
	}

	roles, ok := p.data[authToken]
	if !ok {
		return ok
	}

	if _, ok := roles["admin"]; ok {
		return true
	}

	if _, ok := roles[fmt.Sprintf("ns:%s", namespace)]; !ok {
		return ok
	}
	return true
}
