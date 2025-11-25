package api

import "github.com/gofiber/fiber/v2"

func WebHookTelegram(c *fiber.Ctx) error {

	c.Type("json")

	return c.JSON(fiber.Map{"server": "up"})
}
