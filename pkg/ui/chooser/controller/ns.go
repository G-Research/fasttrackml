package controller

import (
	"github.com/gofiber/fiber/v2"

	repositories "github.com/G-Research/fasttrackml/pkg/api/admin/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/api/admin/service/namespace"

	"github.com/G-Research/fasttrackml/pkg/database"
)

var namespaces []map[string]any

// GetNamespaces renders the index view
func GetNamespaces(ctx *fiber.Ctx) error {
	namespaceService := namespace.NewService(repositories.NewNamespaceRepository(database.DB))
	namespaces, err := namespaceService.ListNamespaces(ctx.Context())
	if err != nil {
		return err
	}
	return ctx.Render("index", fiber.Map{
		"Data": namespaces,
	})
}
