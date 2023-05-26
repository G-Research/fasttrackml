package aim

import (
	"github.com/G-Research/fasttrackml/pkg/api/aim/controller"
	"github.com/gofiber/fiber/v2"
)

func AddRoutes(r fiber.Router) {
	apps := r.Group("apps")
	apps.Get("/", controller.GetApps)
	apps.Post("/", controller.CreateApp)
	apps.Get("/:id/", controller.GetApp)
	apps.Put("/:id/", controller.UpdateApp)
	apps.Delete("/:id/", controller.DeleteApp)

	dashboards := r.Group("/dashboards")
	dashboards.Get("/", controller.GetDashboards)
	dashboards.Post("/", controller.CreateDashboard)
	dashboards.Get("/:id/", controller.GetDashboard)
	dashboards.Put("/:id/", controller.UpdateDashboard)
	dashboards.Delete("/:id/", controller.DeleteDashboard)

	experiments := r.Group("experiments")
	experiments.Get("/", controller.GetExperiments)
	experiments.Get("/:id/", controller.GetExperiment)
	experiments.Get("/:id/activity/", controller.GetExperimentActivity)
	experiments.Get("/:id/runs/", controller.GetExperimentRuns)

	projects := r.Group("/projects")
	projects.Get("/", controller.GetProject)
	projects.Get("/activity/", controller.GetProjectActivity)
	projects.Get("/pinned-sequences/", controller.GetProjectPinnedSequences)
	projects.Post("/pinned-sequences/", controller.UpdateProjectPinnedSequences)
	projects.Get("/params/", controller.GetProjectParams)
	projects.Get("/status/", controller.GetProjectStatus)

	runs := r.Group("/runs")
	runs.Get("/active/", controller.GetRunsActive)
	runs.Get("/search/run/", controller.SearchRuns)
	runs.Get("/search/metric/", controller.SearchMetrics)
	runs.Post("/search/metric/align/", controller.SearchAlignedMetrics)
	runs.Get("/:id/info/", controller.GetRunInfo)
	runs.Post("/:id/metric/get-batch/", controller.GetRunMetrics)

	tags := r.Group("/tags")
	tags.Get("/", controller.GetTags)

	r.Use(func(c *fiber.Ctx) error {
		return fiber.ErrNotFound
	})
}
