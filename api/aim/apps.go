package aim

import "github.com/gofiber/fiber/v2"

func GetApps(c *fiber.Ctx) error {
	return c.JSON([]any{})
}
