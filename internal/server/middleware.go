package server

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(password string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			auth = c.Query("auth")
		}
		if auth != password {
			return c.Status(401).JSON(fiber.Map{
				"timestamp": time.Now().UnixMilli(),
				"status":    401,
				"error":     "Unauthorized",
				"message":   "Invalid authorization",
				"path":      c.Path(),
			})
		}
		return c.Next()
	}
}
