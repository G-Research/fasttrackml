package controller

import (
	"github.com/gofiber/fiber/v2"
)

// NotFoundError renders Not Found error page.
func (c Controller) NotFoundError(ctx *fiber.Ctx) error {
	return ctx.Render("errors/not-found", fiber.Map{})
}
