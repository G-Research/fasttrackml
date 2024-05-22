package controller

import (
	"golang.org/x/oauth2"

	"github.com/G-Research/fasttrackml/pkg/ui/chooser/service/namespace"
)

// Controller handles all the input HTTP requests.
type Controller struct {
	oauth2Config     *oauth2.Config
	namespaceService *namespace.Service
}

// NewController creates a new Controller instance.
func NewController(oauth2Config *oauth2.Config, namespaceService *namespace.Service) *Controller {
	return &Controller{
		oauth2Config:     oauth2Config,
		namespaceService: namespaceService,
	}
}
