package controller

import (
	"github.com/gofiber/fiber/v2"

	"github.com/G-Research/fasttrackml/pkg/ui/admin/request"
	"github.com/G-Research/fasttrackml/pkg/ui/admin/response"
)

// GetNamespaces renders the list view with no message.
func (c Controller) GetNamespaces(ctx *fiber.Ctx) error {
	return c.renderIndex(ctx, "")
}

// GetNamespace renders the update view for a namespace.
func (c Controller) GetNamespace(ctx *fiber.Ctx) error {
	p := struct {
		ID uint `params:"id"`
	}{}
	if err := ctx.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	namespace, err := c.namespaceService.GetNamespace(ctx.Context(), p.ID)
	if err != nil {
		return fiber.NewError(fiber.ErrInternalServerError.Code, "unable to find namespace")
	}
	if namespace == nil {
		return fiber.NewError(fiber.StatusNotFound, "namespace not found")
	}
	return ctx.Render("namespaces/update", fiber.Map{
		"Namespace": namespace,
	})
}

// NewNamespace renders the create view for a namespace.
func (c Controller) NewNamespace(ctx *fiber.Ctx) error {
	namespace := response.Namespace{}
	return ctx.Render("namespaces/create", fiber.Map{
		"Namespace": namespace,
	})
}

// CreateNamespace creates a new namespace record.
func (c Controller) CreateNamespace(ctx *fiber.Ctx) error {
	var namespace request.Namespace
	if err := ctx.BodyParser(&namespace); err != nil {
		return fiber.NewError(400, "unable to parse request body")
	}
	_, err := c.namespaceService.CreateNamespace(ctx.Context(), namespace.Code, namespace.Description)
	if err != nil {
		return ctx.Render("namespaces/create", fiber.Map{
			"Namespace":    namespace,
			"ErrorMessage": err.Error(),
		})
	}
	return c.renderIndex(ctx, "Successfully added new namespace")
}

// UpdateNamespace udpates an existing namespace record.
func (c Controller) UpdateNamespace(ctx *fiber.Ctx) error {
	p := struct {
		ID uint `params:"id"`
	}{}

	if err := ctx.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	var req request.Namespace
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(400, "unable to parse request body")
	}

	_, err := c.namespaceService.UpdateNamespace(ctx.Context(), p.ID, req.Code, req.Description)
	if err != nil {
		return ctx.JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}
	return ctx.JSON(fiber.Map{
		"status":  "success",
		"message": "Successfully updated namespace.",
	})
}

// DeleteNamespace deletes a namespace record.
func (c Controller) DeleteNamespace(ctx *fiber.Ctx) error {
	p := struct {
		ID uint `params:"id"`
	}{}

	if err := ctx.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	err := c.namespaceService.DeleteNamespace(ctx.Context(), p.ID)
	if err != nil {
		return ctx.JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}
	return ctx.JSON(fiber.Map{
		"status":  "success",
		"message": "Successfully deleted namespace.",
	})
}

// renderIndex renders the index page with the given message.
func (c Controller) renderIndex(ctx *fiber.Ctx, msg string) error {
	namespaces, err := c.namespaceService.ListNamespaces(ctx.Context())
	if err != nil {
		return ctx.Render("namespaces/index", fiber.Map{
			"Namespaces":   namespaces,
			"ErrorMessage": err.Error(),
		})
	}
	return ctx.Render("namespaces/index", fiber.Map{
		"Namespaces":     namespaces,
		"SuccessMessage": msg,
	})
}
