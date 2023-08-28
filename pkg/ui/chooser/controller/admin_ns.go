package controller

import (
	"github.com/gofiber/fiber/v2"
)

// AdminNSIndex
func AdminNSIndex(ctx *fiber.Ctx) error {
	return ctx.Render("admin/ns/index", fiber.Map{
		"Title": "Hello, World!",
	})
}
