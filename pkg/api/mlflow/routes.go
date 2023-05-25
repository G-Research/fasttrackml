package mlflow

import (
	"github.com/gofiber/fiber/v2"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/controller"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/metric"
)

// Router represents `mlflow` router.
type Router struct {
	prefixList []string
	controller *controller.Controller
}

// NewRouter creates new instance of `mlflow` router.
func NewRouter(controller *controller.Controller) *Router {
	return &Router{
		prefixList: []string{
			"/api/2.0/mlflow/",
			"/ajax-api/2.0/mlflow/",
			"/mlflow/ajax-api/2.0/mlflow/",
		},
		controller: controller,
	}
}

// Init makes initialization of all `mlflow` routes.
func (r Router) Init(server fiber.Router) {
	for _, prefix := range r.prefixList {
		mainGroup := server.Group(prefix)

		artifacts := mainGroup.Group("/artifacts")
		artifacts.Get("/list", r.controller.ListArtifacts)

		experiments := mainGroup.Group("/experiments")
		experiments.Post("/create", r.controller.CreateExperiment)
		experiments.Post("/delete", r.controller.DeleteExperiment)
		experiments.Get("/get", r.controller.GetExperiment)
		experiments.Get("/get-by-name", r.controller.GetExperimentByName)
		experiments.Get("/list", r.controller.SearchExperiments)
		experiments.Post("/restore", r.controller.RestoreExperiment)
		experiments.Get("/search", r.controller.SearchExperiments)
		experiments.Post("/search", r.controller.SearchExperiments)
		experiments.Post("/set-experiment-tag", r.controller.SetExperimentTag)
		experiments.Post("/update", r.controller.UpdateExperiment)

		metrics := mainGroup.Group("/metrics")
		metrics.Get("/get-history", r.controller.GetMetricHistory)
		metrics.Get("/get-history-bulk", r.controller.GetMetricHistoryBulk)
		metrics.Post("/get-histories", metric.GetMetricHistories)

		runs := mainGroup.Group("/runs")
		runs.Post("/create", r.controller.CreateRun)
		runs.Post("/delete", r.controller.DeleteRun)
		runs.Post("/delete-tag", r.controller.DeleteRunTag)
		runs.Get("/get", r.controller.GetRun)
		runs.Post("/log-batch", r.controller.LogBatch)
		runs.Post("/log-metric", r.controller.LogMetric)
		runs.Post("/log-parameter", r.controller.LogParam)
		runs.Post("/restore", r.controller.RestoreRun)
		runs.Post("/search", r.controller.SearchRuns)
		runs.Post("/set-tag", r.controller.SetRunTag)
		runs.Post("/update", r.controller.UpdateRun)

		mainGroup.Get("/model-versions/search", r.controller.SearchModelVersions)
		mainGroup.Get("/registered-models/search", r.controller.SearchRegisteredModels)

		mainGroup.Use(func(c *fiber.Ctx) error {
			return api.NewEndpointNotFound("Not found")
		})
	}
}
