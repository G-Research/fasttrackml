package aim

import (
	"github.com/gofiber/fiber/v2"
)

func AddRoutes(r fiber.Router) {
	apps := r.Group("apps")
	apps.Get("/", GetApps)
	apps.Post("/", CreateApp)
	apps.Get("/:id/", GetApp)
	apps.Put("/:id/", UpdateApp)
	apps.Delete("/:id/", DeleteApp)

	dashboards := r.Group("/dashboards")
	dashboards.Get("/", GetDashboards)
	dashboards.Post("/", CreateDashboard)
	dashboards.Get("/:id/", GetDashboard)
	dashboards.Put("/:id/", UpdateDashboard)
	dashboards.Delete("/:id/", DeleteDashboard)

	experiments := r.Group("experiments")
	experiments.Get("/", GetExperiments)
	experiments.Get("/:id/", GetExperiment)
	experiments.Get("/:id/activity/", GetExperimentActivity)
	experiments.Get("/:id/runs/", GetExperimentRuns)

	projects := r.Group("/projects")
	projects.Get("/", GetProject)
	projects.Get("/activity/", GetProjectActivity)
	projects.Get("/pinned-sequences/", GetProjectPinnedSequences)
	projects.Post("/pinned-sequences/", UpdateProjectPinnedSequences)
	projects.Get("/params/", GetProjectParams)
	projects.Get("/status/", GetProjectStatus)

	runs := r.Group("/runs")
	runs.Get("/active/", GetRunsActive)
	runs.Get("/search/run/", GetRunsSearch)
	runs.Get("/search/metric/", GetRunsMetricsSearch)
	runs.Get("/:id/info/", GetRunInfo)
	runs.Post("/:id/metric/get-batch/", GetRunMetricBatch)

	tags := r.Group("/tags")
	tags.Get("/", GetTags)

	r.Use(func(c *fiber.Ctx) error {
		return fiber.ErrNotFound
	})
}
