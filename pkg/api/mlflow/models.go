package mlflow

import "github.com/gofiber/fiber/v2"

func SearchModelVersions(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"model_versions": []any{},
	})
}

func SearchRegisteredModels(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"registered_models": []any{},
	})
}
