package chooser

import (
	"embed"
	"github.com/G-Research/fasttrackml/pkg/ui/chooser/controller"
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
	})
	r.Mount("/", app)

	// specific routes
	admin := app.Group("admin")
	admin.Get("ns/new", controller.NewNamespace)
	admin.Get("ns/", controller.GetNamespaces)
	admin.Get("ns/:id/", controller.GetNamespace)
	admin.Put("ns/:id/", controller.PutNamespace)
	admin.Delete("ns/:id/", controller.DeleteNamespace)
	admin.Post("ns/", controller.PostNamespace)


	// default route
	app.Use("/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(sub),
	}))
}
