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
	"github.com/G-Research/fasttrackml/pkg/ui/common"
)

//go:embed embed
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

func (r Router) AddRoutes(fr fiber.Router) {
	sub, _ := fs.Sub(content, "embed")

	fr.Use("/static/chooser/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(sub),
	}))
	// engine and app for template rendering
	engine := html.NewFileSystem(http.FS(sub), ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	fr.Mount("/", app)

	// specific routes
	app.Get("/", r.controller.GetNamespaces)

	// default route
	app.Use("/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(
			common.NewOnlyRootFS(sub, "index.html"),
		),
	}))
}
