package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golink-devs/golink/internal/config"
)

func RegisterRoutes(router fiber.Router, cfg *config.Config) {
	if cfg.Metrics.Enabled {
		RegisterMetrics()
		router.Get("/metrics", MetricsHandler())
	}
}
