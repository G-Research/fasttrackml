package controller

import (
	"github.com/G-Research/fasttrackml/pkg/api/aim/service"
)

// Controller handles all the input HTTP requests.
type Controller struct {
	appService        *service.AppService
	experimentService *service.ExperimentService
}

// NewController creates new Controller instance.
func NewController(appService *service.AppService, experimentService *service.ExperimentService) *Controller {
	return &Controller{
		appService:        appService,
		experimentService: experimentService,
	}
}
