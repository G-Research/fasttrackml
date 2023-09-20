package controller

import (
	"github.com/gofiber/fiber/v2"
)

// GetNamespaces renders the index view
func (c Controller) GetNamespaces(ctx *fiber.Ctx) error {
	// namespaceService := namespace.NewService(repositories.NewNamespaceRepository(database.DB))
	namespaces, err := c.namespaceService.ListNamespaces(ctx.Context())
	if err != nil {
		return err
	}
	return ctx.Render("index", fiber.Map{
		"Data": namespaces,
	})
}
