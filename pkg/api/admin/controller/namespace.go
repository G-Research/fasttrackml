package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"

	"github.com/G-Research/fasttrackml/pkg/api/admin/api/response"
	"github.com/G-Research/fasttrackml/pkg/common/middleware/namespace"
)

// ListNamespaces handles `GET /namespaces/list` endpoint.
func (c Controller) ListNamespaces(ctx *fiber.Ctx) error {
	namespaces, err := c.namespaceService.ListNamespaces(ctx.Context())
	if err != nil {
		return err
	}
	resp := response.NewListNamespacesResponse(namespaces)
	log.Debugf("namespacesList response: %#v", resp)

	return ctx.JSON(resp)
}

// GetCurrentNamespace handles `GET /namespaces/current` endpoint.
func (c Controller) GetCurrentNamespace(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return err
	}
	resp := response.NewGetCurrentNamespaceResponse(ns)
	log.Debugf("currentNamespace response: %#v", resp)

	return ctx.JSON(ns)
}
