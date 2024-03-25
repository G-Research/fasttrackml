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

	"github.com/G-Research/fasttrackml/pkg/api/admin/middleware"
	mlflowConfig "github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
	"github.com/G-Research/fasttrackml/pkg/ui/admin/controller"
)

//go:embed embed/*
var content embed.FS

// Router represents `admin` router.
type Router struct {
	config     *mlflowConfig.ServiceConfig
	controller *controller.Controller
}

// NewRouter creates new instance of `admin` router.
func NewRouter(config *mlflowConfig.ServiceConfig, controller *controller.Controller) *Router {
	return &Router{
		config:     config,
		controller: controller,
	}
}

// Init makes initialization of all `admin` routes.
func (r Router) Init(router fiber.Router) error {
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
	// apply global auth middlewares.
	switch {
	case r.config.Auth.IsAuthTypeUser():
		namespaces.Use(middleware.NewAdminUserMiddleware(r.config.Auth.AuthParsedUserPermissions))
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
