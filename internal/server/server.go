package server

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/golink-devs/golink/api/v1"
	"github.com/golink-devs/golink/api/v4"
	"github.com/golink-devs/golink/internal/config"
	"github.com/golink-devs/golink/internal/hub"
	"github.com/golink-devs/golink/internal/player"
	"github.com/golink-devs/golink/internal/sources"
	"github.com/rs/zerolog/log"
)

type Server struct {
	cfg           *config.Config
	app           *fiber.App
	hub           *hub.Hub
	registry      *sources.Registry
	cache         *sources.TrackCache
	playerManager *player.Manager
}

func New(cfg *config.Config) *Server {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			log.Error().Err(err).Str("path", c.Path()).Int("status", code).Msg("Request error")
			return c.Status(code).JSON(fiber.Map{
				"timestamp": time.Now().UnixMilli(),
				"status":    code,
				"error":     err.Error(),
				"path":      c.Path(),
			})
		},
	})

	app.Use(logger.New())

	h := hub.NewHub()
	playerManager := player.NewManager(h)

	var cache *sources.TrackCache
	if cfg.Cache.Enabled {
		cache = sources.NewTrackCache(time.Duration(cfg.Cache.TTL) * time.Second)
	}
	registry := sources.NewRegistry(cache)

	// Register resolvers
	yt := sources.NewYouTubeResolver()
	if cfg.Sources.YouTube {
		registry.Register(yt)
	}
	if cfg.Sources.Spotify {
		registry.Register(sources.NewSpotifyResolver(cfg.Sources.SpotifyClientID, cfg.Sources.SpotifyClientSecret, yt))
	}
	if cfg.Sources.HTTP {
		registry.Register(sources.NewHTTPResolver())
	}

	s := &Server{
		cfg:           cfg,
		app:           app,
		hub:           h,
		registry:      registry,
		cache:         cache,
		playerManager: playerManager,
	}

	// Root routes
	app.Get("/version", v4.GetVersion)

	// API groups
	v1Group := app.Group("/v1")
	v1.RegisterRoutes(v1Group, cfg)

	v4Group := app.Group("/v4", AuthMiddleware(cfg.Server.Password))
	v4.RegisterRoutes(v4Group, cfg, h, registry, playerManager)

	return s
}

func (s *Server) Start() error {
	go s.hub.Run()

	if s.cache != nil {
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			for range ticker.C {
				s.cache.Cleanup()
			}
		}()
	}

	// Player update ticker
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			active := 0
			playing := 0

			for sessionID, players := range s.playerManager.Sessions() {
				for guildID, p := range players {
					active++
					if p.Track != nil && !p.Paused {
						playing++
						s.hub.Send(sessionID, map[string]interface{}{
							"op":      "playerUpdate",
							"guildId": guildID,
							"state":   p.State(),
						})
					}
				}
			}
			v1.ActivePlayers.Set(float64(active))
			v1.PlayingPlayers.Set(float64(playing))
		}
	}()

	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	log.Info().Str("addr", addr).Msg("Starting GoLink server")
	return s.app.Listen(addr)
}

func (s *Server) Stop() error {
	log.Info().Msg("Stopping GoLink server")
	return s.app.Shutdown()
}
