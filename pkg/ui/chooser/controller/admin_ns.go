package controller

import (
	"github.com/gofiber/fiber/v2"
)

var namespaces []map[string]any

// GetNamespaces renders the index view
func GetNamespaces(ctx *fiber.Ctx) error {
	return ctx.Render("admin/ns/index", fiber.Map{
		"Data": exampleData(), //TODO use service for real data
	})
}

func exampleData() []map[string]any {
	if namespaces == nil {
		namespaces = []map[string]any{
			{"id": 1, "code": "ns1"},
			{"id": 2, "code": "ns2"},
			{"id": 3, "code": "ns3"},
		}
	}
	return namespaces
}
