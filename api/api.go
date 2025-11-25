package api

import (
	"github.com/gofiber/fiber/v2"
)

func InitializeRoute(app *fiber.App) {
	app.Post("/webhook", WebHookTelegram)
	app.Get("/health", HealthHandler)
}
