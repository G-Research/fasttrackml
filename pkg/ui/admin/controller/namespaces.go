package controller

import (
	"github.com/G-Research/fasttrackml/pkg/ui/admin/request"
	"github.com/G-Research/fasttrackml/pkg/ui/admin/response"
	"github.com/gofiber/fiber/v2"
)

// GetNamespaces renders the data for list view.
func (c Controller) GetNamespaces(ctx *fiber.Ctx) error {
	return c.renderIndex(ctx, "")
}

// GetNamespace renders the data for view/edit one namespace
func (c Controller) GetNamespace(ctx *fiber.Ctx) error {
	p := struct {
		ID uint `params:"id"`
	}{}

	if err := ctx.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	ns, err := c.namespaceService.GetNamespace(ctx.Context(), p.ID)
	if err != nil {
		return fiber.NewError(fiber.ErrInternalServerError.Code, "unable to find namespace")
	}
	if ns == nil {
		return fiber.NewError(fiber.StatusNotFound, "namespace not found")
	}

	return ctx.Render("ns/update", fiber.Map{
		"ID":          ns.ID,
		"Code":        ns.Code,
		"Description": ns.Description,
	})
}

// NewNamespace renders the data for view/edit one namespace
func (c Controller) NewNamespace(ctx *fiber.Ctx) error {
	ns := response.Namespace{}
	return ctx.Render("ns/create", fiber.Map{
		"ID":          ns.ID,
		"Code":        ns.Code,
		"Description": ns.Description,
	})
}

// CreateNamespace creates a new namespace record.
func (c Controller) CreateNamespace(ctx *fiber.Ctx) error {
	var req request.Namespace
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(400, "unable to parse request body")
	}
	_, err := c.namespaceService.CreateNamespace(ctx.Context(), req.Code, req.Description)
	if err != nil {
		return ctx.Render("ns/create", fiber.Map{
			"Code":         req.Code,
			"Description":  req.Description,
			"ErrorMessage": err.Error(),
		})
	}

	return c.renderIndex(ctx, "Successfully added new namespace")
}

// UpdateNamespace creates a new namespace record.
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

func (c Controller) renderIndex(ctx *fiber.Ctx, msg string) error {
	namespaces, err := c.namespaceService.ListNamespaces(ctx.Context())
	if err != nil {
		return ctx.Render("ns/index", fiber.Map{
			"Data":         namespaces,
			"ErrorMessage": err.Error(),
		})
	}
	return ctx.Render("ns/index", fiber.Map{
		"Data":           namespaces,
		"SuccessMessage": msg,
	})
}
