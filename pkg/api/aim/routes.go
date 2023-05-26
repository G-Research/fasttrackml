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

func (rtr Router) AddRoutes(r fiber.Router) {
	apps := r.Group("apps")
	apps.Get("/", rtr.controller.GetApps)
	apps.Post("/", rtr.controller.CreateApp)
	apps.Get("/:id/", rtr.controller.GetApp)
	apps.Put("/:id/", rtr.controller.UpdateApp)
	apps.Delete("/:id/", rtr.controller.DeleteApp)

	dashboards := r.Group("/dashboards")
	dashboards.Get("/", rtr.controller.GetDashboards)
	dashboards.Post("/", rtr.controller.CreateDashboard)
	dashboards.Get("/:id/", rtr.controller.GetDashboard)
	dashboards.Put("/:id/", rtr.controller.UpdateDashboard)
	dashboards.Delete("/:id/", rtr.controller.DeleteDashboard)

	experiments := r.Group("experiments")
	experiments.Get("/", rtr.controller.GetExperiments)
	experiments.Get("/:id/", rtr.controller.GetExperiment)
	experiments.Get("/:id/activity/", rtr.controller.GetExperimentActivity)
	experiments.Get("/:id/runs/", rtr.controller.GetExperimentRuns)

	projects := r.Group("/projects")
	projects.Get("/", rtr.controller.GetProject)
	projects.Get("/activity/", rtr.controller.GetProjectActivity)
	projects.Get("/pinned-sequences/", rtr.controller.GetProjectPinnedSequences)
	projects.Post("/pinned-sequences/", rtr.controller.UpdateProjectPinnedSequences)
	projects.Get("/params/", rtr.controller.GetProjectParams)
	projects.Get("/status/", rtr.controller.GetProjectStatus)

	runs := r.Group("/runs")
	runs.Get("/active/", rtr.controller.GetRunsActive)
	runs.Get("/search/run/", rtr.controller.SearchRuns)
	runs.Get("/search/metric/", rtr.controller.SearchMetrics)
	runs.Post("/search/metric/align/", rtr.controller.SearchAlignedMetrics)
	runs.Get("/:id/info/", rtr.controller.GetRunInfo)
	runs.Post("/:id/metric/get-batch/", rtr.controller.GetRunMetrics)

	tags := r.Group("/tags")
	tags.Get("/", rtr.controller.GetTags)
	r.Use(func(c *fiber.Ctx) error {
		return fiber.ErrNotFound
	})
}
