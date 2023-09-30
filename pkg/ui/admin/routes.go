package admin

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"

	"github.com/G-Research/fasttrackml/pkg/ui/admin/controller"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
)

//go:embed embed/*
var content embed.FS

// Router represents `mlflow` router.
type Router struct {
	controller *controller.Controller
}

// NewRouter creates new instance of `mlflow` router.
func NewRouter(controller *controller.Controller) *Router {
	return &Router{
		controller: controller,
	}
}

// Init makes initialization of all `mlflow` routes.
func (r Router) Init(fr fiber.Router) {
	sub, _ := fs.Sub(content, "embed")

	// engine and app for template rendering
	engine := html.NewFileSystem(http.FS(sub), ".html")
	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layouts/main",
	})
	fr.Mount("/admin", app)

	// specific routes
	namespaces := app.Group("ns")
	namespaces.Get("/", r.controller.GetNamespaces)
	namespaces.Post("/", r.controller.CreateNamespace)
	namespaces.Get("/new", r.controller.NewNamespace)
	namespaces.Get("/:id/", r.controller.GetNamespace)
	namespaces.Put("/:id/", r.controller.UpdateNamespace)
	namespaces.Delete("/:id/", r.controller.DeleteNamespace)

	// default route
	app.Use("/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(sub),
	}))
}
