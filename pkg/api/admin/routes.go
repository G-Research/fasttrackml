package admin

import (
	"github.com/gofiber/fiber/v2"

	"github.com/G-Research/fasttrackml/pkg/api/admin/controller"
)

// Router represents `admin` router.
type Router struct {
	controller *controller.Controller
}

// NewRouter creates new instance of `admin` router.
func NewRouter(controller *controller.Controller) *Router {
	return &Router{
		controller: controller,
	}
}

// Init makes initialization of all `admin` routes.
func (r Router) Init(router fiber.Router) {
	mainGroup := router.Group("admin")
	namespaces := mainGroup.Group("namespaces")
	namespaces.Get("/list", r.controller.ListNamespaces)
	namespaces.Get("/current", r.controller.GetCurrentNamespace)
}
