package mlflow

import (
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/controller"
	"github.com/gofiber/fiber/v2"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/artifact"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/experiment"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/metric"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/model"
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
		artifacts.Get("/list", artifact.ListArtifacts)

		experiments := mainGroup.Group("/experiments")
		experiments.Post("/create", experiment.CreateExperiment)
		experiments.Post("/delete", experiment.DeleteExperiment)
		experiments.Get("/get", experiment.GetExperiment)
		experiments.Get("/get-by-name", experiment.GetExperimentByName)
		experiments.Get("/list", experiment.SearchExperiments)
		experiments.Post("/restore", experiment.RestoreExperiment)
		experiments.Get("/search", experiment.SearchExperiments)
		experiments.Post("/search", experiment.SearchExperiments)
		experiments.Post("/set-experiment-tag", experiment.SetExperimentTag)
		experiments.Post("/update", experiment.UpdateExperiment)

		metrics := mainGroup.Group("/metrics")
		metrics.Get("/get-history", metric.GetMetricHistory)
		metrics.Get("/get-history-bulk", metric.GetMetricHistoryBulk)
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

		mainGroup.Get("/model-versions/search", model.SearchModelVersions)
		mainGroup.Get("/registered-models/search", model.SearchRegisteredModels)

		mainGroup.Use(func(c *fiber.Ctx) error {
			return api.NewEndpointNotFound("Not found")
		})
	}
}
