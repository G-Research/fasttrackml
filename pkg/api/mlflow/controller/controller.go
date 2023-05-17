package controller

import (
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/run"
)

// Controller handles all the input HTTP requests.
type Controller struct {
	runService *run.Service
}

// NewController creates new Controller instance.
func NewController(runService *run.Service) *Controller {
	return &Controller{
		runService: runService,
	}
}
