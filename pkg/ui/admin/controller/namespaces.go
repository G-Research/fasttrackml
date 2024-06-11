package controller

import (
	"github.com/gofiber/fiber/v2"

	"github.com/G-Research/fasttrackml/pkg/ui/admin/request"
	"github.com/G-Research/fasttrackml/pkg/ui/admin/response"
	"github.com/G-Research/fasttrackml/pkg/ui/common"
)

const (
	StatusError   = "error"
	StatusSuccess = "success"
)

// GetNamespaces renders the list view with no message.
func (c Controller) GetNamespaces(ctx *fiber.Ctx) error {
	return c.renderIndex(ctx, "")
}

// GetNamespace renders the update view for a namespace.
func (c Controller) GetNamespace(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "unable to parse id")
	}
	namespace, err := c.namespaceService.GetNamespace(ctx.Context(), uint(id))
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
			"Namespace": namespace,
			"Status":    StatusError,
			"Message":   common.ErrorMessageForUI("namespace code", err.Error()),
		})
	}
	return c.renderIndex(ctx, "Successfully added new namespace")
}

// UpdateNamespace updates an existing namespace record.
func (c Controller) UpdateNamespace(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "unable to parse id")
	}
	var req request.Namespace
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(400, "unable to parse request body")
	}

	_, err = c.namespaceService.UpdateNamespace(ctx.Context(), uint(id), req.Code, req.Description)
	if err != nil {
		return ctx.JSON(fiber.Map{
			"status":  StatusError,
			"message": common.ErrorMessageForUI("namespace code", err.Error()),
		})
	}
	return ctx.JSON(fiber.Map{
		"status":  StatusSuccess,
		"message": "Successfully updated namespace.",
	})
}

// DeleteNamespace deletes a namespace record.
func (c Controller) DeleteNamespace(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "unable to parse id")
	}
	err = c.namespaceService.DeleteNamespace(ctx.Context(), uint(id))
	if err != nil {
		return ctx.JSON(fiber.Map{
			"status":  StatusError,
			"message": common.ErrorMessageForUI("namespace code", err.Error()),
		})
	}
	return ctx.JSON(fiber.Map{
		"status":  StatusSuccess,
		"message": "Successfully deleted namespace.",
	})
}

// renderIndex renders the index page with the given message.
func (c Controller) renderIndex(ctx *fiber.Ctx, msg string) error {
	namespaces, err := c.namespaceService.ListNamespaces(ctx.Context())
	if err != nil {
		return ctx.Render("namespaces/index", fiber.Map{
			"Namespaces": namespaces,
			"Status":     StatusError,
			"Message":    common.ErrorMessageForUI("namespace", err.Error()),
		})
	}
	return ctx.Render("namespaces/index", fiber.Map{
		"Namespaces": namespaces,
		"Status":     StatusSuccess,
		"Message":    msg,
	})
}
