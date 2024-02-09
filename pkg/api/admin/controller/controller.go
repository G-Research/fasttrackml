package controller

import "github.com/G-Research/fasttrackml/pkg/api/admin/service/namespace"

// Controller contains all the request handler functions for the admin api.
type Controller struct {
	namespaceService *namespace.Service
}

// NewController creates new Controller instance.
func NewController(namespaceService *namespace.Service) *Controller {
	return &Controller{
		namespaceService: namespaceService,
	}
}
