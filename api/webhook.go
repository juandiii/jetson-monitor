package api

import (
	"github.com/gofiber/fiber"
)

func InitializeRoute(app *fiber.App) {
	app.Post("/webhook", WebHookTelegram)
}

func WebHookTelegram(c *fiber.Ctx) {

	c.Type("json")

	c.JSON(fiber.Map{"server": "up"})
}
