package mlflow

import (
	"github.com/gofiber/fiber/v2"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/controller"
	"github.com/G-Research/fasttrackml/pkg/common/api"
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
	ArtifactsGetRoute  = "/get"
	ArtifactsListRoute = "/list"
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
	RunsGetRoute          = "/get"
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
	RunsLogOutputRoute    = "/log-output"
	RunsLogArtifactRoute  = "/log-artifact"
)

// Router represents `mlflow` router.
type Router struct {
	prefixList        []string
	controller        *controller.Controller
	globalMiddlewares []fiber.Handler
}

// NewRouter creates new instance of `mlflow` router.
func NewRouter(controller *controller.Controller) *Router {
	return &Router{
		prefixList: []string{
			"/api/2.0/mlflow/",
			"/ajax-api/2.0/mlflow/",
		},
		controller:        controller,
		globalMiddlewares: make([]fiber.Handler, 0),
	}
}

// Init makes initialization of all `mlflow` routes.
func (r *Router) Init(router fiber.Router) {
	for _, prefix := range r.prefixList {
		mainGroup := router.Group(prefix)
		// apply global middlewares.
		for _, globalMiddleware := range r.globalMiddlewares {
			mainGroup.Use(globalMiddleware)
		}

		// setup related routes.
		artifacts := mainGroup.Group(ArtifactsRoutePrefix)
		artifacts.Get(ArtifactsGetRoute, r.controller.GetArtifact)
		artifacts.Get(ArtifactsListRoute, r.controller.ListArtifacts)

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
		runs.Get(RunsGetRoute, r.controller.GetRun)
		runs.Post(RunsLogBatchRoute, r.controller.LogBatch)
		runs.Post(RunsLogMetricRoute, r.controller.LogMetric)
		runs.Post(RunsLogParameterRoute, r.controller.LogParam)
		runs.Post(RunsRestoreRoute, r.controller.RestoreRun)
		runs.Post(RunsSearchRoute, r.controller.SearchRuns)
		runs.Post(RunsSetTagRoute, r.controller.SetRunTag)
		runs.Post(RunsUpdateRoute, r.controller.UpdateRun)
		runs.Post(RunsLogOutputRoute, r.controller.LogOutput)
		runs.Post(RunsLogArtifactRoute, r.controller.LogArtifact)

		mainGroup.Get("/model-versions/search", r.controller.SearchModelVersions)
		mainGroup.Get("/registered-models/search", r.controller.SearchRegisteredModels)

		mainGroup.Use(func(c *fiber.Ctx) error {
			return api.NewEndpointNotFound("Not found")
		})
	}
}

// AddGlobalMiddleware adds a global middleware which will be applied for each route.
func (r *Router) AddGlobalMiddleware(middleware fiber.Handler) *Router {
	r.globalMiddlewares = append(r.globalMiddlewares, middleware)
	return r
}
