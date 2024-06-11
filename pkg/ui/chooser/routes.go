package chooser

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"
	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/ui/chooser/controller"
)

//go:embed embed
var content embed.FS

// Router represents `chooser` router.
type Router struct {
	controller        *controller.Controller
	globalMiddlewares []fiber.Handler
}

// NewRouter creates new instance of `chooser` router.
func NewRouter(controller *controller.Controller) *Router {
	return &Router{
		controller: controller,
	}
}

// Init adds all the `chooser` routes
func (r *Router) Init(router fiber.Router) error {
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

	// apply global middlewares.
	for _, globalMiddleware := range r.globalMiddlewares {
		app.Use(globalMiddleware)
	}

	// setup related routes.
	app.Get("/login", r.controller.Login)

	app.Get("/", r.controller.GetNamespaces)
	app.Get("/chooser/namespaces", r.controller.ListNamespaces)
	app.Get("/chooser/namespaces/current", r.controller.GetCurrentNamespace)

	// setup routes to static files.
	app.Use("/chooser/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(sub),
	}))

	errors := app.Group("errors")
	errors.Get("/not-found", r.controller.NotFoundError)
	errors.Get("/internal-server", r.controller.InternalServerError)

	return nil
}

// AddGlobalMiddleware adds a global middleware which will be applied for each route.
func (r *Router) AddGlobalMiddleware(middleware fiber.Handler) *Router {
	r.globalMiddlewares = append(r.globalMiddlewares, middleware)
	return r
}
