package admin

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"
	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/ui/admin/controller"
)

//go:embed embed/*
var content embed.FS

// Router represents `admin` router.
type Router struct {
	controller        *controller.Controller
	globalMiddlewares []fiber.Handler
}

// NewRouter creates new instance of `admin` router.
func NewRouter(controller *controller.Controller) *Router {
	return &Router{
		controller: controller,
	}
}

// Init makes initialization of all `admin` routes.
func (r *Router) Init(router fiber.Router) error {
	//nolint:errcheck
	sub, err := fs.Sub(content, "embed")
	if err != nil {
		return eris.Wrap(err, "error mounting `embed` directory")
	}

	// engine and app for template rendering
	engine := html.NewFileSystem(http.FS(sub), ".html")
	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layouts/main",
	})
	router.Mount("/admin", app)

	// specific routes
	namespaces := app.Group("namespaces")
	// apply global middlewares.
	for _, globalMiddleware := range r.globalMiddlewares {
		namespaces.Use(globalMiddleware)
	}
	namespaces.Get("/", r.controller.GetNamespaces)
	namespaces.Post("/", r.controller.CreateNamespace)
	namespaces.Get("/new", r.controller.NewNamespace)
	namespaces.Get("/:id<int>/", r.controller.GetNamespace)
	namespaces.Put("/:id<int>/", r.controller.UpdateNamespace)
	namespaces.Delete("/:id<int>/", r.controller.DeleteNamespace)

	// default route
	app.Use("/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(sub),
	}))

	return nil
}

// AddGlobalMiddleware adds a global middleware which will be applied for each route.
func (r *Router) AddGlobalMiddleware(middleware fiber.Handler) *Router {
	r.globalMiddlewares = append(r.globalMiddlewares, middleware)
	return r
}
