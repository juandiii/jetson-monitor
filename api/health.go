package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/juandiii/jetson-monitor/version"
)

func HealthHandler(c *fiber.Ctx) error {
	c.Type("json")
	v, commit, build := version.Info()
	return c.JSON(fiber.Map{"status": "ok", "version": v, "commit": commit, "build_time": build})
}
