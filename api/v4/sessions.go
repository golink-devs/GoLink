package v4

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golink-devs/golink/internal/hub"
)

type UpdateSessionRequest struct {
	Resuming *bool `json:"resuming,omitempty"`
	Timeout  *int  `json:"timeout,omitempty"`
}

func RegisterSessionRoutes(router fiber.Router, h *hub.Hub) {
	router.Patch("/sessions/:sessionId", UpdateSession(h))
}

func UpdateSession(h *hub.Hub) fiber.Handler {
	return func(c *fiber.Ctx) error {
		_ = c.Params("sessionId")
		var req UpdateSessionRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		// TODO: Implement session resuming logic in Hub
		log := h.Send // just to use h for now
		_ = log

		return c.JSON(req)
	}
}
