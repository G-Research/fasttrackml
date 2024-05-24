package controller

import (
	"github.com/gofiber/fiber/v2"
)

// Login renders Login page.
func (c *Controller) Login(ctx *fiber.Ctx) error {
	if c.config.Auth.IsAuthTypeOIDC() {
		return ctx.Render("login/login", fiber.Map{
			"authUrl": c.oidcClient.GetOauth2Config().AuthCodeURL(
				GenerateRandomString(20),
			),
		})
	}
	return ctx.Redirect("/")
}
