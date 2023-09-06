package admin

import (
	"embed"
	"github.com/G-Research/fasttrackml/pkg/ui/admin/controller"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"
	"io/fs"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
)

//go:embed embed/*
var content embed.FS

func AddRoutes(r fiber.Router) {
	sub, _ := fs.Sub(content, "embed")

	// engine and app for template rendering
	engine := html.NewFileSystem(http.FS(sub), ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
		ViewsLayout: "layouts/main",
	})
	r.Mount("/", app)

	// specific routes
	namespaces := app.Group("ns")
	namespaces.Get("/", controller.GetNamespaces)
	namespaces.Post("/", controller.CreateNamespace)
	namespaces.Get("/new", controller.NewNamespace)
	namespaces.Get("/:id/", controller.GetNamespace)
	namespaces.Put("/:id/", controller.UpdateNamespace)
	namespaces.Delete("/:id/", controller.DeleteNamespace)

	// default route
	app.Use("/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(sub),
	}))
}
