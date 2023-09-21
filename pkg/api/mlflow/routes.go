package mlflow

import (
	"github.com/gofiber/fiber/v2"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/controller"
)

// List of route prefixes.
const (
	RunsRoutePrefix        = "/runs"
	MetricsRoutePrefix     = "/metrics"
	ArtifactsRoutePrefix   = "/artifacts"
	ExperimentsRoutePrefix = "/experiments"
)

// List of `/artifact/*` routes.
const (
	ArtifactsListRoute = "/list"
	ArtifactsGetRoute  = "/get-artifact"
)

// List of `/experiments/*` routes.
const (
	ExperimentsGetRoute         = "/get"
	ExperimentsListRoute        = "/list"
	ExperimentsCreateRoute      = "/create"
	ExperimentsDeleteRoute      = "/delete"
	ExperimentsRestoreRoute     = "/restore"
	ExperimentsSearchRoute      = "/search"
	ExperimentsUpdateRoute      = "/update"
	ExperimentsGetByNameRoute   = "/get-by-name"
	ExperimentsSetExperimentTag = "/set-experiment-tag"
)

// List of `/metrics/*` routes.
const (
	MetricsGetHistoriesRoute   = "/get-histories"
	MetricsGetHistoryRoute     = "/get-history"
	MetricsGetHistoryBulkRoute = "/get-history-bulk"
)

// List of `/runs/*` routes.
const (
	RunsGetTagRoute       = "/get"
	RunsCreateRoute       = "/create"
	RunsDeleteRoute       = "/delete"
	RunsSearchRoute       = "/search"
	RunsSetTagRoute       = "/set-tag"
	RunsUpdateRoute       = "/update"
	RunsRestoreRoute      = "/restore"
	RunsDeleteTagRoute    = "/delete-tag"
	RunsLogBatchRoute     = "/log-batch"
	RunsLogMetricRoute    = "/log-metric"
	RunsLogParameterRoute = "/log-parameter"
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

		artifacts := mainGroup.Group(ArtifactsRoutePrefix)
		artifacts.Get(ArtifactsListRoute, r.controller.ListArtifacts)
		artifacts.Get(ArtifactsGetRoute, r.controller.GetArtifact)

		experiments := mainGroup.Group(ExperimentsRoutePrefix)
		experiments.Post(ExperimentsCreateRoute, r.controller.CreateExperiment)
		experiments.Post(ExperimentsDeleteRoute, r.controller.DeleteExperiment)
		experiments.Get(ExperimentsGetRoute, r.controller.GetExperiment)
		experiments.Get(ExperimentsGetByNameRoute, r.controller.GetExperimentByName)
		experiments.Get(ExperimentsListRoute, r.controller.SearchExperiments)
		experiments.Post(ExperimentsRestoreRoute, r.controller.RestoreExperiment)
		experiments.Get(ExperimentsSearchRoute, r.controller.SearchExperiments)
		experiments.Post(ExperimentsSearchRoute, r.controller.SearchExperiments)
		experiments.Post(ExperimentsSetExperimentTag, r.controller.SetExperimentTag)
		experiments.Post(ExperimentsUpdateRoute, r.controller.UpdateExperiment)

		metrics := mainGroup.Group(MetricsRoutePrefix)
		metrics.Get(MetricsGetHistoryRoute, r.controller.GetMetricHistory)
		metrics.Get(MetricsGetHistoryBulkRoute, r.controller.GetMetricHistoryBulk)
		metrics.Post(MetricsGetHistoriesRoute, r.controller.GetMetricHistories)

		runs := mainGroup.Group(RunsRoutePrefix)
		runs.Post(RunsCreateRoute, r.controller.CreateRun)
		runs.Post(RunsDeleteRoute, r.controller.DeleteRun)
		runs.Post(RunsDeleteTagRoute, r.controller.DeleteRunTag)
		runs.Get(RunsGetTagRoute, r.controller.GetRun)
		runs.Post(RunsLogBatchRoute, r.controller.LogBatch)
		runs.Post(RunsLogMetricRoute, r.controller.LogMetric)
		runs.Post(RunsLogParameterRoute, r.controller.LogParam)
		runs.Post(RunsRestoreRoute, r.controller.RestoreRun)
		runs.Post(RunsSearchRoute, r.controller.SearchRuns)
		runs.Post(RunsSetTagRoute, r.controller.SetRunTag)
		runs.Post(RunsUpdateRoute, r.controller.UpdateRun)

		mainGroup.Get("/model-versions/search", r.controller.SearchModelVersions)
		mainGroup.Get("/registered-models/search", r.controller.SearchRegisteredModels)

		mainGroup.Use(func(c *fiber.Ctx) error {
			return api.NewEndpointNotFound("Not found")
		})
	}
}
