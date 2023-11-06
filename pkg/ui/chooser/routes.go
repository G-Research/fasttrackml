package chooser

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	namespaceMiddleware "github.com/G-Research/fasttrackml/pkg/common/middleware/namespace"
	"github.com/G-Research/fasttrackml/pkg/ui/chooser/controller"
)

//go:embed embed
var content embed.FS

// Router represents `chooser` router.
type Router struct {
	controller          *controller.Controller
	namespaceRepository repositories.NamespaceRepositoryProvider
}

// NewRouter creates new instance of `chooser` router.
func NewRouter(
	controller *controller.Controller,
	namespaceRepository repositories.NamespaceRepositoryProvider,
) *Router {
	return &Router{
		controller:          controller,
		namespaceRepository: namespaceRepository,
	}
}

// AddRoutes adds all the `chooser` routes
func (r Router) AddRoutes(fr fiber.Router) {
	//nolint:errcheck
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
	app.Get("/", namespaceMiddleware.New(r.namespaceRepository), r.controller.GetNamespaces)
}
