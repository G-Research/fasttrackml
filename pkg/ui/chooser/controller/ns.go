package controller

import (
	"github.com/gofiber/fiber/v2"
	//"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/namespace"
)

var namespaces []map[string]any

// GetNamespaces renders the index view
func GetNamespaces(ctx *fiber.Ctx) error {
	/*var namespaceService namespace.Service
	namespaces, err := namespaceService.ListNamespaces(ctx.Context())
	if err != nil {
		return err
	}
	*/
	return ctx.Render("index", fiber.Map{
		"Data": exampleData(),
	})
}

// exampleData TODO remove this, used for UI dev
func exampleData() []map[string]any {
	if namespaces == nil {
		namespaces = []map[string]any{
			{"ID": 1, "code": "Namespace1"},
			{"ID": 2, "code": "Namespace2"},
			{"ID": 3, "code": "Namespace3"},
			{"ID": 4, "code": "Namespace4"},
			{"ID": 5, "code": "Namespace5"},
			{"ID": 6, "code": "Namespace6"},
			{"ID": 7, "code": "Namespace7"},
			{"ID": 8, "code": "Namespace8"},
			{"ID": 9, "code": "Namespace9"},
			{"ID": 10, "code": "Namespace10"},
		}
	}
	return namespaces
}
