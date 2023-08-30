package controller

import (
	"github.com/gofiber/fiber/v2"
)

var namespaces []map[string]any

// GetNamespaces renders the index view
func GetNamespaces(ctx *fiber.Ctx) error {
	return ctx.Render("admin/ns/index", fiber.Map{
		"Data": exampleData(), //TODO use service for real data
		"ErrorMessage": "",
		"SuccessMessage": "",
	})
}

// TODO this is just for UI dev
func exampleData() []map[string]any {
	if namespaces == nil {
		namespaces = []map[string]any{
			{"id": 1, "code": "ns1", "description": "This is namespace 1", "created_at": "2023-08-28"},
			{"id": 2, "code": "ns2", "description": "This is namespace 2", "created_at": "2023-08-29"},
			{"id": 3, "code": "ns3", "description": "This is namespace 3", "created_at": "2023-08-30"},
		}
	}
	return namespaces
}
