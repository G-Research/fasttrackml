package controller

import "github.com/gofiber/fiber/v2"

// Login renders Login page.
func (c Controller) Login(ctx *fiber.Ctx) error {
	return ctx.Render("login/login", fiber.Map{})
}
