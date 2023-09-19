package mlflow

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/filesystem"

	mlflowUI "github.com/G-Research/fasttrackml-ui-mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/controller"
	"github.com/G-Research/fasttrackml/pkg/ui/common"
)

// Router represents `mlflow` UI router.
type Router struct {
	controller *controller.Controller
}

// NewRouter creates new instance of `mlflow` router.
func NewRouter(controller *controller.Controller) *Router {
	return &Router{
		controller: controller,
	}
}

// Init configures the UI routes
func (router Router) Init(r fiber.Router) {
	// Handle MLFlow requests for artifacts
	r.Get("/mlflow/get-artifact", router.controller.GetArtifact) 
	
	r.Use("/static/mlflow/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(mlflowUI.FS),
	}))

	r.Use("/mlflow/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(
			common.NewOnlyRootFS(mlflowUI.FS, "index.html"),
		),
	}))
}
