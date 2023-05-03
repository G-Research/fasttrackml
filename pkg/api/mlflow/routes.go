package mlflow

import (
	"github.com/G-Research/fasttrack/pkg/api/mlflow/api"
	"github.com/gofiber/fiber/v2"
)

func AddRoutes(r fiber.Router) {
	artifacts := r.Group("/artifacts")
	artifacts.Get("/list", ListArtifacts)

	experiments := r.Group("/experiments")
	experiments.Post("/create", CreateExperiment)
	experiments.Post("/delete", DeleteExperiment)
	experiments.Get("/get", GetExperiment)
	experiments.Get("/get-by-name", GetExperimentByName)
	experiments.Get("/list", SearchExperiments)
	experiments.Post("/restore", RestoreExperiment)
	experiments.Get("/search", SearchExperiments)
	experiments.Post("/search", SearchExperiments)
	experiments.Post("/set-experiment-tag", SetExperimentTag)
	experiments.Post("/update", UpdateExperiment)

	metrics := r.Group("/metrics")
	metrics.Get("/get-history", GetMetricHistory)
	metrics.Get("/get-history-bulk", GetMetricHistoryBulk)
	metrics.Post("/get-histories", GetMetricHistories)

	runs := r.Group("/runs")
	runs.Post("/create", CreateRun)
	runs.Post("/delete", DeleteRun)
	runs.Post("/delete-tag", DeleteRunTag)
	runs.Get("/get", GetRun)
	runs.Post("/log-batch", LogBatch)
	runs.Post("/log-metric", LogMetric)
	runs.Post("/log-parameter", LogParam)
	runs.Post("/restore", RestoreRun)
	runs.Post("/search", SearchRuns)
	runs.Post("/set-tag", SetRunTag)
	runs.Post("/update", UpdateRun)

	r.Get("/model-versions/search", SearchModelVersions)
	r.Get("/registered-models/search", SearchRegisteredModels)

	r.Use(func(c *fiber.Ctx) error {
		return api.NewEndpointNotFound("Not found")
	})
}
