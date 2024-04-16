package aim2

import (
	"github.com/gofiber/fiber/v2"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/controller"
	"github.com/G-Research/fasttrackml/pkg/common/config"
	"github.com/G-Research/fasttrackml/pkg/common/middleware"
)

// Router represents `mlflow` router.
type Router struct {
	config     *config.Config
	controller *controller.Controller
}

// NewRouter creates new instance of `mlflow` router.
func NewRouter(config *config.Config, controller *controller.Controller) *Router {
	return &Router{
		config:     config,
		controller: controller,
	}
}

func (r Router) Init(server fiber.Router) {
	mainGroup := server.Group("/aim/api")
	// apply global auth middlewares.
	switch {
	case r.config.Auth.IsAuthTypeUser():
		mainGroup.Use(middleware.NewUserMiddleware(r.config.Auth.AuthParsedUserPermissions))
	}

	// setup related routes.
	apps := mainGroup.Group("apps")
	apps.Get("/", r.controller.GetApps)
	apps.Post("/", r.controller.CreateApp)
	apps.Get("/:id/", r.controller.GetApp)
	apps.Put("/:id/", r.controller.UpdateApp)
	apps.Delete("/:id/", r.controller.DeleteApp)

	dashboards := mainGroup.Group("/dashboards")
	dashboards.Get("/", r.controller.GetDashboards)
	dashboards.Post("/", r.controller.CreateDashboard)
	dashboards.Get("/:id/", r.controller.GetDashboard)
	dashboards.Put("/:id/", r.controller.UpdateDashboard)
	dashboards.Delete("/:id/", r.controller.DeleteDashboard)

	experiments := mainGroup.Group("experiments")
	experiments.Get("/", r.controller.GetExperiments)
	experiments.Get("/:id/", r.controller.GetExperiment)
	experiments.Get("/:id/activity/", r.controller.GetExperimentActivity)
	experiments.Get("/:id/runs/", r.controller.GetExperimentRuns)
	experiments.Delete("/:id/", r.controller.DeleteExperiment)
	experiments.Put("/:id/", r.controller.UpdateExperiment)

	projects := mainGroup.Group("/projects")
	projects.Get("/", r.controller.GetProject)
	projects.Get("/activity/", r.controller.GetProjectActivity)
	projects.Get("/pinned-sequences/", r.controller.GetProjectPinnedSequences)
	projects.Post("/pinned-sequences/", r.controller.UpdateProjectPinnedSequences)
	projects.Get("/params/", r.controller.GetProjectParams)
	projects.Get("/status/", r.controller.GetProjectStatus)

	runs := mainGroup.Group("/runs")
	runs.Get("/active/", r.controller.GetRunsActive)
	runs.Get("/search/run/", r.controller.SearchRuns)
	runs.Post("/search/metric/", r.controller.SearchMetrics)
	runs.Post("/search/metric/align/", r.controller.SearchAlignedMetrics)
	runs.Get("/:id/info/", r.controller.GetRunInfo)
	runs.Post("/:id/metric/get-batch/", r.controller.GetRunMetrics)
	runs.Put("/:id/", r.controller.UpdateRun)
	runs.Delete("/:id/", r.controller.DeleteRun)
	runs.Post("/delete-batch/", r.controller.DeleteBatch)
	runs.Post("/archive-batch/", r.controller.ArchiveBatch)

	tags := mainGroup.Group("/tags")
	tags.Get("/", r.controller.GetTags)

	mainGroup.Use(func(c *fiber.Ctx) error {
		return fiber.ErrNotFound
	})
}
