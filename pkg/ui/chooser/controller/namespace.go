package controller

import (
	"github.com/gofiber/fiber/v2"

	commonMiddleware "github.com/G-Research/fasttrackml/pkg/common/middleware"
)

// GetNamespaces renders the index view
func (c Controller) GetNamespaces(ctx *fiber.Ctx) error {
	ns, err := commonMiddleware.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return err
	}

	namespaces, isAdmin, err := c.namespaceService.ListNamespaces(ctx.Context())
	if err != nil {
		return err
	}

	return ctx.Render("index", fiber.Map{
		"IsAdmin":          isAdmin,
		"Namespaces":       namespaces,
		"CurrentNamespace": ns,
	})
}
