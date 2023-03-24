package aim

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
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			var e *ErrorResponse
			var f *fiber.Error
			var d DetailedError

			switch {
			case errors.As(err, &e):
			case errors.As(err, &f):
				e = &ErrorResponse{
					Code:    f.Code,
					Message: f.Message,
					Detail:  "",
				}
			case errors.As(err, &d):
				e = &ErrorResponse{
					Code:    d.Code(),
					Message: d.Message(),
					Detail:  d.Detail(),
				}
			default:
				e = &ErrorResponse{
					Code:    fiber.StatusInternalServerError,
					Message: err.Error(),
					Detail:  "",
				}
			}

			fn := log.Errorf

			switch e.Code {
			case fiber.StatusNotFound:
				fn = log.Debugf
			case fiber.StatusInternalServerError:
			default:
				fn = log.Warnf
			}

			fn("Error encountered in %s %s: %s", c.Method(), c.Path(), err)

			return c.Status(e.Code).JSON(e)
		},
	})

	api := app.Group("/api")

	if authUsername != "" && authPassword != "" {
		log.Infof(`BasicAuth enabled for modern UI with user "%s"`, authUsername)
		api.Use(basicauth.New(basicauth.Config{
			Users: map[string]string{
				authUsername: authPassword,
			},
		}))
	}

	apps := api.Group("apps")
	apps.Get("/", GetApps)
	apps.Post("/", CreateApp)
	apps.Get("/:id/", GetApp)
	apps.Put("/:id/", UpdateApp)
	apps.Delete("/:id/", DeleteApp)

	dashboards := api.Group("/dashboards")
	dashboards.Get("/", GetDashboards)
	dashboards.Post("/", CreateDashboard)
	dashboards.Get("/:id/", GetDashboard)
	dashboards.Put("/:id/", UpdateDashboard)
	dashboards.Delete("/:id/", DeleteDashboard)

	experiments := api.Group("experiments")
	experiments.Get("/", GetExperiments)
	experiments.Get("/:id/", GetExperiment)
	experiments.Get("/:id/activity/", GetExperimentActivity)
	experiments.Get("/:id/runs/", GetExperimentRuns)

	projects := api.Group("/projects")
	projects.Get("/", GetProject)
	projects.Get("/activity/", GetProjectActivity)
	projects.Get("/pinned-sequences/", GetProjectPinnedSequences)
	projects.Post("/pinned-sequences/", UpdateProjectPinnedSequences)
	projects.Get("/params/", GetProjectParams)
	projects.Get("/status/", GetProjectStatus)

	runs := api.Group("/runs")
	runs.Get("/active/", GetRunsActive)
	runs.Get("/search/run/", GetRunsSearch)
	runs.Get("/search/metric/", GetRunsMetricsSearch)
	runs.Get("/:id/info/", GetRunInfo)
	runs.Post("/:id/metric/get-batch/", GetRunMetricBatch)

	tags := api.Group("/tags")
	tags.Get("/", GetTags)

	api.Get("/version", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"version": version.Version,
		})
	})

	api.Use(func(c *fiber.Ctx) error {
		return fiber.ErrNotFound
	})

	app.Use("/static-files/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(ui.AimFS),
	}))

	app.Use(etag.New(), func(c *fiber.Ctx) error {
		if c.Method() != fiber.MethodGet {
			return fiber.ErrMethodNotAllowed
		}

		file, _ := ui.AimFS.Open("index.html")
		stat, _ := file.Stat()
		c.Set("Content-Type", "text/html; charset=utf-8")
		c.Response().SetBodyStream(file, int(stat.Size()))
		return nil
	})

	return app
}
