package aim

import (
	"errors"
	"net/http"

	"github.com/G-Resarch/fasttrack/ui"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	log "github.com/sirupsen/logrus"
)

func NewApp(authUsername string, authPassword string) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Status code defaults to 500
			code := fiber.StatusInternalServerError
			fn := log.Errorf

			// Retrieve the custom status code if it's a *fiber.Error
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
				switch e.Code {
				case fiber.StatusNotFound:
					fn = log.Debugf
				case fiber.StatusInternalServerError:
				default:
					fn = log.Warnf
				}
			}

			fn("Error encountered in %s %s: %s", c.Method(), c.Path(), err)

			return c.Status(code).JSON(fiber.Map{
				"detail": err.Error(),
			})
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

	dashboards := api.Group("/dashboards")
	dashboards.Get("/", GetDashboards)

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
