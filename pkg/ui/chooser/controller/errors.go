package controller

import (
	"github.com/gofiber/fiber/v2"
)

// NotFoundError renders Not Found error page.
func (c Controller) NotFoundError(ctx *fiber.Ctx) error {
	return ctx.Render("errors/not-found", fiber.Map{})
}

// InternalServerError renders Internal Server error page.
func (c Controller) InternalServerError(ctx *fiber.Ctx) error {
	return ctx.Render("errors/internal-server-error", fiber.Map{})
}
