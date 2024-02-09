package controller

import "github.com/gofiber/fiber/v2"

func (c Controller) GetTags(ctx *fiber.Ctx) error {
	return ctx.JSON([]string{})
}
