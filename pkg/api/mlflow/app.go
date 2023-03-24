package mlflow

import (
	"errors"
	"net/http"

	"github.com/G-Resarch/fasttrack/pkg/ui"
	"github.com/G-Resarch/fasttrack/pkg/version"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	log "github.com/sirupsen/logrus"
)

func NewApp(authUsername string, authPassword string) *fiber.App {
	api := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			var e *ErrorResponse
			if !errors.As(err, &e) {
				var code ErrorCode = ErrorCodeInternalError

				var f *fiber.Error
				if errors.As(err, &f) {
					switch f.Code {
					case fiber.StatusBadRequest:
						code = ErrorCodeBadRequest
					case fiber.StatusServiceUnavailable:
						code = ErrorCodeTemporarilyUnavailable
					case fiber.StatusNotFound:
						code = ErrorCodeEndpointNotFound
					}
				}

				e = &ErrorResponse{
					ErrorCode: code,
					Message:   err.Error(),
				}
			}

			var code int
			var fn func(format string, args ...any)

			switch e.ErrorCode {
			case ErrorCodeBadRequest, ErrorCodeInvalidParameterValue, ErrorCodeResourceAlreadyExists:
				code = fiber.StatusBadRequest
				fn = log.Infof
			case ErrorCodeTemporarilyUnavailable:
				code = fiber.StatusServiceUnavailable
				fn = log.Warnf
			case ErrorCodeEndpointNotFound, ErrorCodeResourceDoesNotExist:
				code = fiber.StatusNotFound
				fn = log.Debugf
			default:
				code = fiber.StatusInternalServerError
				fn = log.Errorf
			}

			fn("Error encountered in %s %s: %s", c.Method(), c.Path(), err)

			return c.Status(code).JSON(e)
		},
	})

	if authUsername != "" && authPassword != "" {
		log.Infof(`BasicAuth enabled for classic UI with user "%s"`, authUsername)
		api.Use(basicauth.New(basicauth.Config{
			Users: map[string]string{
				authUsername: authPassword,
			},
		}))
	}

	artifacts := api.Group("/artifacts")
	artifacts.Get("/list", ListArtifacts)

	experiments := api.Group("/experiments")
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

	metrics := api.Group("/metrics")
	metrics.Get("/get-history", GetMetricHistory)
	metrics.Get("/get-history-bulk", GetMetricHistoryBulk)
	metrics.Post("/get-histories", GetMetricHistories)

	runs := api.Group("/runs")
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

	api.Get("/model-versions/search", SearchModelVersions)
	api.Get("/registered-models/search", SearchRegisteredModels)

	api.Use(func(c *fiber.Ctx) error {
		return NewError(ErrorCodeEndpointNotFound, "Not found")
	})

	app := fiber.New()

	app.Mount("/api/2.0/mlflow", api)
	app.Mount("/ajax-api/2.0/mlflow/", api)
	app.Mount("/api/2.0/preview/mlflow/", api)
	app.Mount("/ajax-api/2.0/preview/mlflow/", api)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Get("/version", func(c *fiber.Ctx) error {
		return c.SendString(version.Version)
	})

	app.Use("/static-files/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(ui.MlflowFS),
	}))

	app.Get("/", etag.New(), func(c *fiber.Ctx) error {
		file, _ := ui.MlflowFS.Open("index.html")
		stat, _ := file.Stat()
		c.Set("Content-Type", "text/html; charset=utf-8")
		c.Response().SetBodyStream(file, int(stat.Size()))
		return nil
	})

	return app
}
