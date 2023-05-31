package controller

import (
	"github.com/G-Research/fasttrackml/pkg/service/artifact"
	"github.com/G-Research/fasttrackml/pkg/service/experiment"
	"github.com/G-Research/fasttrackml/pkg/service/metric"
	"github.com/G-Research/fasttrackml/pkg/service/model"
	"github.com/G-Research/fasttrackml/pkg/service/run"
)

// Controller handles all the input HTTP requests.
type Controller struct {
	appService        *app.Service
	runService        *run.Service
	modelService      *model.Service
	metricService     *metric.Service
	artifactService   *artifact.Service
	experimentService *experiment.Service
}

// NewController creates new Controller instance.
func NewController(runService *run.Service, modelService *model.Service, metricService *metric.Service, artifactService *artifact.Service, experimentService *experiment.Service, appService *app.Service) *Controller {
	return &Controller{
		appService:        appService,
		modelService:      modelService,
		metricService:     metricService,
		artifactService:   artifactService,
		experimentService: experimentService,
	}
}
