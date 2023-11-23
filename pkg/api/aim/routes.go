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
	experiments.Delete("/:id/", DeleteExperiment)
	experiments.Put("/:id/", UpdateExperiment)

	projects := r.Group("/projects")
	projects.Get("/", GetProject)
	projects.Get("/activity/", GetProjectActivity)
	projects.Get("/pinned-sequences/", GetProjectPinnedSequences)
	projects.Post("/pinned-sequences/", UpdateProjectPinnedSequences)
	projects.Get("/params/", GetProjectParams)
	projects.Get("/status/", GetProjectStatus)

	runs := r.Group("/runs")
	runs.Get("/active/", GetRunsActive)
	runs.Get("/search/run/", SearchRuns)
	runs.Get("/search/metric/", SearchMetrics)
	runs.Post("/search/metric/align/", SearchAlignedMetrics)
	runs.Get("/:id/info/", GetRunInfo)
	runs.Post("/:id/metric/get-batch/", GetRunMetrics)
	runs.Put("/:id/", UpdateRun)
	runs.Delete("/:id/", DeleteRun)
	runs.Post("/delete-batch/", DeleteBatch)
	runs.Post("/archive-batch/", ArchiveBatch)

	tags := r.Group("/tags")
	tags.Get("/", GetTags)

	r.Use(func(c *fiber.Ctx) error {
		return fiber.ErrNotFound
	})
}
