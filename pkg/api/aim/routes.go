package aim

import (
	"github.com/G-Research/fasttrackml/pkg/api/aim/controller"
	"github.com/gofiber/fiber/v2"
)

// Router represents `aim` router.
type Router struct {
	controller *controller.Controller
}

// NewRouter creates new instance of `mlflow` router.
func NewRouter(controller *controller.Controller) *Router {
	return &Router{
		controller: controller,
	}
}

func (r Router) AddRoutes(server fiber.Router) {
	apps := server.Group("apps")
	apps.Get("/", r.controller.GetApps)
	apps.Post("/", r.controller.CreateApp)
	apps.Get("/:id/", r.controller.GetApp)
	apps.Put("/:id/", r.controller.UpdateApp)
	apps.Delete("/:id/", r.controller.DeleteApp)

	dashboards := server.Group("/dashboards")
	dashboards.Get("/", r.controller.GetDashboards)
	dashboards.Post("/", r.controller.CreateDashboard)
	dashboards.Get("/:id/", r.controller.GetDashboard)
	dashboards.Put("/:id/", r.controller.UpdateDashboard)
	dashboards.Delete("/:id/", r.controller.DeleteDashboard)

	experiments := server.Group("experiments")
	experiments.Get("/", r.controller.GetExperiments)
	experiments.Get("/:id/", r.controller.GetExperiment)
	experiments.Get("/:id/activity/", r.controller.GetExperimentActivity)
	experiments.Get("/:id/runs/", r.controller.GetExperimentRuns)

	projects := server.Group("/projects")
	projects.Get("/", r.controller.GetProject)
	projects.Get("/activity/", r.controller.GetProjectActivity)
	projects.Get("/pinned-sequences/", r.controller.GetProjectPinnedSequences)
	projects.Post("/pinned-sequences/", r.controller.UpdateProjectPinnedSequences)
	projects.Get("/params/", r.controller.GetProjectParams)
	projects.Get("/status/", r.controller.GetProjectStatus)

	runs := server.Group("/runs")
	runs.Get("/active/", r.controller.GetRunsActive)
	runs.Get("/search/run/", r.controller.SearchRuns)
	runs.Get("/search/metric/", r.controller.SearchMetrics)
	runs.Post("/search/metric/align/", r.controller.SearchAlignedMetrics)
	runs.Get("/:id/info/", r.controller.GetRunInfo)
	runs.Post("/:id/metric/get-batch/", r.controller.GetRunMetrics)
	runs.Put("/:id/", r.controller.UpdateRun)
	runs.Delete("/:id/", r.controller.DeleteRun)
	runs.Post("/delete-batch/", r.controller.DeleteBatch)
	runs.Post("/archive-batch/", r.controller.ArchiveBatch)

	tags := server.Group("/tags")
	tags.Get("/", r.controller.GetTags)
	server.Use(func(c *fiber.Ctx) error {
		return fiber.ErrNotFound
	})
}
