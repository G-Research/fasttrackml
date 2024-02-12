package controller

import (
	"github.com/G-Research/fasttrackml/pkg/api/aim2/service"
)

// Controller handles all the input HTTP requests.
type Controller struct {
	svc *service.Service
}

// NewController creates new Controller instance.
func NewController(svc *service.Service) *Controller {
	return &Controller{svc: svc}
}
