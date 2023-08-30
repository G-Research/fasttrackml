package controller

import (
	"github.com/gofiber/fiber/v2"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/namespace"
)

// GetNamespaces renders the index view
func GetNamespaces(ctx *fiber.Ctx) error {
	var namespaceService namespace.Service
	namespaces, err := namespaceService.ListNamespaces(ctx.Context())
	if err != nil {
		return err
	}
	return ctx.Render("ns/index", fiber.Map{
		"Data": namespaces,
	})
}
