package aim

import "github.com/gofiber/fiber/v2"

func GetTags(c *fiber.Ctx) error {
	return c.JSON([]string{})
}
