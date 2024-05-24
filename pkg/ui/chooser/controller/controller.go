package controller

import (
	"github.com/G-Research/fasttrackml/pkg/common/auth/oidc"
	"github.com/G-Research/fasttrackml/pkg/common/config"
	"github.com/G-Research/fasttrackml/pkg/ui/chooser/service/namespace"
)

// Controller handles all the input HTTP requests.
type Controller struct {
	config           *config.Config
	oidcClient       oidc.ClientProvider
	namespaceService *namespace.Service
}

// NewController creates a new Controller instance.
func NewController(config *config.Config, namespaceService *namespace.Service) *Controller {
	return &Controller{
		config:           config,
		namespaceService: namespaceService,
	}
}

// SetOIDCClient sets OIDC client.
func (c *Controller) SetOIDCClient(oidcClient oidc.ClientProvider) {
	c.oidcClient = oidcClient
}
