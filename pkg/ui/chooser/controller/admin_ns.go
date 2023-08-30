package controller

import (
	"time"

	"github.com/G-Research/fasttrackml/pkg/api/chooser/request"
	"github.com/G-Research/fasttrackml/pkg/api/chooser/response"
	"github.com/gofiber/fiber/v2"
)

var namespaces []response.Namespace

// GetNamespaces renders the data for list view.
func GetNamespaces(ctx *fiber.Ctx) error {
	return ctx.Render("admin/ns/index", fiber.Map{
		"Data": exampleData(), //TODO use service for real data
		"ErrorMessage": "",
		"SuccessMessage": "",
	})
}

// PostNamespace creates a new namespace record.
func PostNamespace(ctx *fiber.Ctx) error {
	var req request.CreateNamespace
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(400, "unable to parse request body")
	}
	addNamespace(response.Namespace{
		Code: req.Code,
		Description: req.Code,
	})
	return ctx.Render("admin/ns/index", fiber.Map{
		"Data": exampleData(), //TODO use service for real data
		"ErrorMessage": "",
		"SuccessMessage": "Successfully added new namespace",
	})
}

// PutNamespace creates a new namespace record.
func PutNamespace(ctx *fiber.Ctx) error {
	var req request.CreateNamespace
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(400, "unable to parse request body")
	}
	addNamespace(response.Namespace{
		Code: req.Code,
		Description: req.Code,
	})
	return ctx.Render("admin/ns/index", fiber.Map{
		"Data": exampleData(), //TODO use service for real data
		"ErrorMessage": "",
		"SuccessMessage": "Successfully added new namespace",
	})
}

// exampleData TODO remove this, used for UI dev 
func exampleData() []response.Namespace {
	if namespaces == nil {
		namespaces = []response.Namespace{
			{ ID: 1, Code: "ns1", Description: "This is namespace 1", CreatedAt: time.Now()},
			{ ID: 2, Code: "ns2", Description: "This is namespace 2", CreatedAt: time.Now()},
			{ ID: 3, Code: "ns3", Description: "This is namespace 3", CreatedAt: time.Now()},
			{ ID: 4, Code: "ns4", Description: "This is namespace 4", CreatedAt: time.Now()},
		}
	}
	return namespaces
}

// findNamespace TODO remove this, used for UI dev 
func findNamespace(id uint) *response.Namespace {
	for _, ns := range namespaces {
		if ns.ID == id {
			return &ns
		}
	}
	return nil
}

// addNamespace TODO remove this, used for UI dev 
func addNamespace(newNS response.Namespace) {
	newNS.ID = namespaces[len(namespaces) - 1].ID + 1
	newNS.CreatedAt = time.Now()
	namespaces = append(namespaces, newNS)
}

// deleteNamespace TODO remove this, used for UI dev 
func deleteNamespace(id uint) {
	ns := findNamespace(id)
	if ns != nil {
		deletedAt := time.Now()
		ns.DeletedAt = &deletedAt
	}
	return
}

