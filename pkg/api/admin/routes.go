package admin

import (
	"github.com/gofiber/fiber/v2"

	"github.com/G-Research/fasttrackml/pkg/api/admin/controller"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
)

// List of route prefixes.
const (
	NamespacesRoutePrefix = "/ns"
)

// List of `/namespaces/*` routes.
const (
	NamespacesListRoute = "/list"
)

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
func (r Router) Init(server fiber.Router) {
	mainGroup := server.Group("/admin")

	namespaces := mainGroup.Group(NamespacesRoutePrefix)
	namespaces.Get(NamespacesListRoute, r.controller.ListNamespaces)

	mainGroup.Use(func(c *fiber.Ctx) error {
		return api.NewEndpointNotFound("Not found")
	})

}
