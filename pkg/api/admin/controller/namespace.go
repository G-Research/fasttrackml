package controller

import (
	"github.com/gofiber/fiber/v2"

	"github.com/G-Research/fasttrackml/pkg/common/middleware/namespace"
)

// ListNamespaces handles `GET /namespaces/list` endpoint.
func (c Controller) ListNamespaces(ctx *fiber.Ctx) error {
	namespaces, err := c.namespaceService.ListNamespaces(ctx.Context())
	if err != nil {
		return err
	}

	return ctx.JSON(namespaces)
}

// GetCurrentNamespace handles `GET /namespaces/current` endpoint.
func (c Controller) GetCurrentNamespace(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return err
	}

	return ctx.JSON(ns)
}
