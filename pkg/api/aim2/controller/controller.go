package controller

import (
	"github.com/G-Research/fasttrackml/pkg/api/aim2/service/app"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/service/dashboard"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/service/experiment"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/service/project"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/service/run"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/service/tag"
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
