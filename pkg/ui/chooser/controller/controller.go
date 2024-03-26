package controller

import "github.com/G-Research/fasttrackml/pkg/ui/chooser/service/namespace"

// Controller handles all the input HTTP requests.
type Controller struct {
	namespaceService *namespace.Service
}

// NewController creates new Controller instance.
func NewController(namespaceService *namespace.Service) *Controller {
	return &Controller{
		namespaceService: namespaceService,
	}
}
