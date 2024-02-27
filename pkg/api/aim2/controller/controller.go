package controller

import (
	"github.com/G-Research/fasttrackml/pkg/api/aim2/services/app"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/services/dashboard"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/services/experiment"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/services/project"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/services/run"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/services/tag"
)

// Controller handles all the input HTTP requests.
type Controller struct {
	tagService        *tag.Service
	appService        *app.Service
	runService        *run.Service
	projectService    *project.Service
	dashboardService  *dashboard.Service
	experimentService *experiment.Service
}

// NewController creates new Controller instance.
func NewController(
	tagService *tag.Service,
	appService *app.Service,
	runService *run.Service,
	projectService *project.Service,
	dashboardService *dashboard.Service,
	experimentService *experiment.Service,
) *Controller {
	return &Controller{
		tagService:        tagService,
		appService:        appService,
		runService:        runService,
		projectService:    projectService,
		dashboardService:  dashboardService,
		experimentService: experimentService,
	}
}
