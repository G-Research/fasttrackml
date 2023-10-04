package chooser

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"

	"github.com/G-Research/fasttrackml/pkg/ui/chooser/controller"
)

//go:embed embed
var content embed.FS

type Router struct {
	controller *controller.Controller
}

// NewRouter creates new instance of `chooser` router.
func NewRouter(controller *controller.Controller) *Router {
	return &Router{
		controller: controller,
	}
}

// AddRoutes adds all the `chooser` routes
func (r Router) AddRoutes(fr fiber.Router) {
	sub, _ := fs.Sub(content, "embed")

	fr.Use("/static/chooser/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(sub),
	}))

	// app for template rendering
	app := fiber.New(fiber.Config{
		Views: html.NewFileSystem(http.FS(sub), ".html"),
	})
	fr.Mount("/", app)

	// specific routes
	app.Get("/", r.controller.GetNamespaces)
}
