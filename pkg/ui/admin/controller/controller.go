package controller

import "github.com/G-Research/fasttrackml/pkg/ui/admin/service/namespace"

// Controller contains all the request handler functions for the admin ui.
type Controller struct {
	namespaceService *namespace.Service
}

// NewController creates new Controller instance.
func NewController(namespaceService *namespace.Service) *Controller {
	return &Controller{
		namespaceService: namespaceService,
	}
}
