package v4

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/golink-devs/golink/internal/config"
	"github.com/golink-devs/golink/internal/hub"
	"github.com/golink-devs/golink/internal/player"
	"github.com/golink-devs/golink/internal/sources"
)

func RegisterRoutes(router fiber.Router, cfg *config.Config, h *hub.Hub, registry *sources.Registry, manager *player.Manager) {
	// Info routes
	router.Get("/info", GetInfo)
	router.Get("/version", GetVersion)
	router.Get("/stats", GetStats)

	// Session routes
	RegisterSessionRoutes(router, h)

	// WebSocket route
	router.Get("/websocket", websocket.New(NewWebSocketHandler(h)))

	// Track routes
	RegisterTrackRoutes(router, registry)

	// Player routes
	RegisterPlayerRoutes(router, manager, h, registry)

	// Other modules will register their routes here
}
