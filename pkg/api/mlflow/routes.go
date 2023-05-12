package mlflow

import (
	"github.com/gofiber/fiber/v2"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/artifact"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/experiment"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/metric"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/model"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/run"
)

func AddRoutes(r fiber.Router) {
	artifacts := r.Group("/artifacts")
	artifacts.Get("/list", artifact.ListArtifacts)

	experiments := r.Group("/experiments")
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

	metrics := r.Group("/metrics")
	metrics.Get("/get-history", metric.GetMetricHistory)
	metrics.Get("/get-history-bulk", metric.GetMetricHistoryBulk)
	metrics.Post("/get-histories", metric.GetMetricHistories)

	runs := r.Group("/runs")
	runs.Post("/create", run.CreateRun)
	runs.Post("/delete", run.DeleteRun)
	runs.Post("/delete-tag", run.DeleteRunTag)
	runs.Get("/get", run.GetRun)
	runs.Post("/log-batch", run.LogBatch)
	runs.Post("/log-metric", run.LogMetric)
	runs.Post("/log-parameter", run.LogParam)
	runs.Post("/restore", run.RestoreRun)
	runs.Post("/search", run.SearchRuns)
	runs.Post("/set-tag", run.SetRunTag)
	runs.Post("/update", run.UpdateRun)

	r.Get("/model-versions/search", model.SearchModelVersions)
	r.Get("/registered-models/search", model.SearchRegisteredModels)

	r.Use(func(c *fiber.Ctx) error {
		return api.NewEndpointNotFound("Not found")
	})
}
