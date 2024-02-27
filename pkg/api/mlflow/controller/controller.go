package controller

import (
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/services/artifact"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/services/experiment"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/services/metric"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/services/model"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/services/run"
)

// Controller handles all the input HTTP requests.
type Controller struct {
	runService        *run.Service
	modelService      *model.Service
	metricService     *metric.Service
	artifactService   *artifact.Service
	experimentService *experiment.Service
}

// NewController creates new Controller instance.
func NewController(
	runService *run.Service,
	modelService *model.Service,
	metricService *metric.Service,
	artifactService *artifact.Service,
	experimentService *experiment.Service,
) *Controller {
	return &Controller{
		runService:        runService,
		modelService:      modelService,
		metricService:     metricService,
		artifactService:   artifactService,
		experimentService: experimentService,
	}
}
