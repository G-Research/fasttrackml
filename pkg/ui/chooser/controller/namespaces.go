package controller

import (
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/common/middleware"
	"github.com/G-Research/fasttrackml/pkg/ui/chooser/api/response"
)

// GetNamespaces renders the index view
func (c *Controller) GetNamespaces(ctx *fiber.Ctx) error {
	namespaces, isAdmin, err := c.namespaceService.ListNamespaces(ctx.Context())
	if err != nil {
		return err
	}
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return err
	}
	return ctx.Render("namespaces/index", fiber.Map{
		"IsAdmin":          isAdmin,
		"Namespaces":       namespaces,
		"CurrentNamespace": ns,
	})
}

// ListNamespaces handles `GET /namespaces` endpoint.
func (c *Controller) ListNamespaces(ctx *fiber.Ctx) error {
	namespaces, _, err := c.namespaceService.ListNamespaces(ctx.Context())
	if err != nil {
		return err
	}
	resp := response.NewListNamespacesResponse(namespaces)
	log.Debugf("namespacesList response: %#v", resp)

	return ctx.JSON(resp)
}

// GetCurrentNamespace handles `GET /namespaces/current` endpoint.
func (c *Controller) GetCurrentNamespace(ctx *fiber.Ctx) error {
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return err
	}
	resp := response.NewGetCurrentNamespaceResponse(ns)
	log.Debugf("currentNamespace response: %#v", resp)

	return ctx.JSON(ns)
}
