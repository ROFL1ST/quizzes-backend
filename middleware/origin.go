package middleware

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/gofiber/fiber/v2"
)

func EnforceOriginForStateChanging() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() == fiber.MethodOptions {
			return c.Next()
		}

		switch c.Method() {
		case fiber.MethodPost, fiber.MethodPut, fiber.MethodPatch, fiber.MethodDelete:
			origin := c.Get("Origin")
			if origin != "" && !config.IsAllowedOrigin(origin) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"status":  "error",
					"message": "Invalid request origin",
				})
			}
		}

		return c.Next()
	}
}
