package controller

import (
	"github.com/G-Research/fasttrackml/pkg/api/aim/service"
)

// Controller handles all the input HTTP requests.
type Controller struct {
	appService        *AppService
}

// NewController creates new Controller instance.
func NewController(
	appService *AppService,
) *Controller {
	return &Controller{
		appService:        appService,
	}
}
