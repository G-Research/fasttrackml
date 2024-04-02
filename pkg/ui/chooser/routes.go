package chooser

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/filesystem"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/rotisserie/eris"

	mlflowConfig "github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
	"github.com/G-Research/fasttrackml/pkg/ui/chooser/controller"
	"github.com/G-Research/fasttrackml/pkg/ui/chooser/middleware"
)

//go:embed embed
var content embed.FS

// Router represents `chooser` router.
type Router struct {
	config     *mlflowConfig.ServiceConfig
	controller *controller.Controller
}

// NewRouter creates new instance of `chooser` router.
func NewRouter(config *mlflowConfig.ServiceConfig, controller *controller.Controller) *Router {
	return &Router{
		config:     config,
		controller: controller,
	}
}

// Init adds all the `chooser` routes
func (r Router) Init(router fiber.Router) error {
	//nolint:errcheck
	sub, err := fs.Sub(content, "embed")
	if err != nil {
		return eris.Wrap(err, "error mounting `embed` directory")
	}

	// app for template rendering
	app := fiber.New(fiber.Config{
		Views:       html.NewFileSystem(http.FS(sub), ".html"),
		ViewsLayout: "layouts/main",
	})
	router.Mount("/", app)

	// apply global auth middlewares.
	switch {
	case r.config.Auth.IsAuthTypeUser():
		app.Use(middleware.NewUserMiddleware(r.config.Auth.AuthParsedUserPermissions))
	}

	// setup related routes.
	app.Get("/", r.controller.GetNamespaces)
	app.Get("/chooser/namespaces", r.controller.ListNamespaces)
	app.Get("/chooser/namespaces/current", r.controller.GetCurrentNamespace)

	// setup routes to static files.
	app.Use("/chooser/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(sub),
	}))

	errors := app.Group("errors")
	errors.Get("/not-found", r.controller.NotFoundError)

	return nil
}
