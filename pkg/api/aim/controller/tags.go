package controller

import "github.com/gofiber/fiber/v2"

func (ctlr Controller) GetTags(c *fiber.Ctx) error {
	return c.JSON([]string{})
}
