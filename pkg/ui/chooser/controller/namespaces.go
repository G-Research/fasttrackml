package controller

import (
	"github.com/gofiber/fiber/v2"

	"github.com/G-Research/fasttrackml/pkg/common/middleware"
)

// GetNamespaces renders the index view
func (c Controller) GetNamespaces(ctx *fiber.Ctx) error {
	namespaces, isAdmin, err := c.namespaceService.ListNamespaces(ctx.Context())
	if err != nil {
		return err
	}
	ns, err := middleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return err
	}
	return ctx.Render("index", fiber.Map{
		"IsAdmin":          isAdmin,
		"Namespaces":       namespaces,
		"CurrentNamespace": ns,
	})
}
