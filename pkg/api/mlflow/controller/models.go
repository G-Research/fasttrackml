package controller

import "github.com/gofiber/fiber/v2"

// SearchModelVersions handles `GET /model-versions/search` endpoint.
func (c Controller) SearchModelVersions(ctx *fiber.Ctx) error {
	models, err := c.modelService.SearchModelVersions(ctx.Context())
	if err != nil {
		return err
	}
	return ctx.JSON(models)
}

// SearchRegisteredModels handles `GET /registered-models/search` endpoint.
func (c Controller) SearchRegisteredModels(ctx *fiber.Ctx) error {
	models, err := c.modelService.SearchRegisteredModels(ctx.Context())
	if err != nil {
		return err
	}
	return ctx.JSON(models)
}
