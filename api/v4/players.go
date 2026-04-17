package v4

import (
	"context"
	"encoding/json"
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/golink-devs/golink/internal/hub"
	"github.com/golink-devs/golink/internal/player"
	"github.com/golink-devs/golink/internal/sources"
	"github.com/golink-devs/golink/internal/voice"
	"github.com/rs/zerolog/log"
)

type PlayerResponse struct {
	GuildID string             `json:"guildId"`
	Track   *sources.Track     `json:"track"`
	Volume  int                `json:"volume"`
	Paused  bool               `json:"paused"`
	State   player.PlayerState `json:"state"`
	Voice   *VoiceState        `json:"voice"`
	Filters player.Filters     `json:"filters"`
}

type VoiceState struct {
	Token     string `json:"token"`
	Endpoint  string `json:"endpoint"`
	SessionID string `json:"sessionId"`
}

type UpdatePlayerRequest struct {
	Track      json.RawMessage `json:"track,omitempty"`
	Identifier string          `json:"identifier,omitempty"`
	Position   *int64          `json:"position,omitempty"`
	EndTime    *int64          `json:"endTime,omitempty"`
	Volume     *int            `json:"volume,omitempty"`
	Paused     *bool           `json:"paused,omitempty"`
	Filters    *player.Filters `json:"filters,omitempty"`
	Voice      *VoiceState     `json:"voice,omitempty"`
}

type UpdateTrackRequest struct {
	Encoded  string                 `json:"encoded,omitempty"`
	UserData map[string]interface{} `json:"userData,omitempty"`
}

func RegisterPlayerRoutes(router fiber.Router, manager *player.Manager, h *hub.Hub, registry *sources.Registry) {
	router.Get("/sessions/:sessionId/players", GetPlayers(manager))
	router.Get("/sessions/:sessionId/players/:guildId", GetPlayer(manager))
	router.Patch("/sessions/:sessionId/players/:guildId", UpdatePlayer(manager, h, registry))
	router.Delete("/sessions/:sessionId/players/:guildId", DeletePlayer(manager))
}

func GetPlayers(manager *player.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Params("sessionId")
		players := manager.GetPlayers(sessionID)
		if players == nil {
			return c.JSON([]interface{}{})
		}

		resp := make([]PlayerResponse, 0, len(players))
		for _, p := range players {
			resp = append(resp, mapPlayerToResponse(p))
		}
		return c.JSON(resp)
	}
}

func GetPlayer(manager *player.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Params("sessionId")
		guildID := c.Params("guildId")
		p, ok := manager.GetPlayer(sessionID, guildID)
		if !ok {
			return c.Status(404).JSON(fiber.Map{
				"timestamp": time.Now().UnixMilli(),
				"status":    404,
				"error":     "Not Found",
				"message":   "Player not found",
				"path":      c.Path(),
			})
		}
		return c.JSON(mapPlayerToResponse(p))
	}
}

func UpdatePlayer(manager *player.Manager, h *hub.Hub, registry *sources.Registry) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Params("sessionId")
		guildID := c.Params("guildId")
		noReplace := c.Query("noReplace") == "true"

		var req UpdatePlayerRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"timestamp": time.Now().UnixMilli(),
				"status":    400,
				"error":     "Bad Request",
				"message":   err.Error(),
				"path":      c.Path(),
			})
		}

		p, ok := manager.GetPlayer(sessionID, guildID)
		if !ok {
			client, ok := h.GetClient(sessionID)
			if !ok {
				log.Warn().Str("session_id", sessionID).Msg("Session not found during player update")
				return c.Status(400).JSON(fiber.Map{
					"timestamp": time.Now().UnixMilli(),
					"status":    400,
					"error":     "Bad Request",
					"message":   "Session not found",
					"path":      c.Path(),
				})
			}
			p = manager.CreatePlayer(sessionID, client.UserID, guildID)
		}

		if noReplace && p.Track != nil {
			return c.JSON(mapPlayerToResponse(p))
		}

		// Handle track update
		var track *sources.Track
		if len(req.Track) > 0 {
			var encoded string
			// Try to unmarshal as object first
			var utr UpdateTrackRequest
			if err := json.Unmarshal(req.Track, &utr); err == nil && utr.Encoded != "" {
				encoded = utr.Encoded
			} else {
				// Fallback to unmarshalling as plain string
				if err := json.Unmarshal(req.Track, &encoded); err != nil {
					log.Debug().Err(err).Str("track_raw", string(req.Track)).Msg("Failed to unmarshal track as string or object")
				}
			}

			if encoded != "" {
				info, err := sources.DecodeTrack(encoded)
				if err == nil {
					track = &sources.Track{
						Encoded: encoded,
						Info:    *info,
					}
				} else {
					log.Debug().Err(err).Str("encoded", encoded).Msg("Failed to decode track info")
				}
			}
		} else if req.Identifier != "" {
			res, err := registry.Resolve(c.Context(), req.Identifier)
			if err == nil && res.LoadType == sources.LoadTypeTrack {
				t := res.Data.(sources.Track)
				track = &t
			}
		}

		if track != nil {
			var streamURL string
			if track.Info.SourceName == "youtube" {
				yt := sources.NewYouTubeResolver()
				streamURL, _ = yt.GetStreamURL(c.Context(), track.Info.URI)
			} else if track.Info.SourceName == "spotify" {
				yt := sources.NewYouTubeResolver()
				sp := sources.NewSpotifyResolver("", "", yt)
				res, err := sp.ResolveToYouTube(c.Context(), track.Info.Title, track.Info.Author)
				if err == nil && res.LoadType == sources.LoadTypeTrack {
					t := res.Data.(sources.Track)
					streamURL, _ = yt.GetStreamURL(c.Context(), t.Info.URI)
				}
			} else if track.Info.SourceName == "http" {
				streamURL = track.Info.URI
			}

			if streamURL != "" {
				// Use context.Background() so playback continues after request
				if err := p.Play(context.Background(), *track, streamURL); err != nil {
					log.Error().Err(err).Msg("Failed to start playback")
				}
			}
		}

		if req.Volume != nil {
			p.SetVolume(*req.Volume)
		}
		if req.Paused != nil {
			p.Pause(*req.Paused)
		}
		if req.Filters != nil {
			p.SetFilters(*req.Filters)
		}
		if req.Position != nil {
			p.Seek(*req.Position)
		}

		if req.Voice != nil {
			userID, _ := snowflake.Parse(p.UserID)
			guildIDID, _ := snowflake.Parse(p.GuildID)
			if p.VoiceConn == nil {
				p.VoiceConn = voice.NewVoiceConn(userID, guildIDID)
			}
			// Voice connection opening would happen here in a full implementation
		}

		return c.JSON(mapPlayerToResponse(p))
	}
}

func DeletePlayer(manager *player.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Params("sessionId")
		guildID := c.Params("guildId")
		p, ok := manager.GetPlayer(sessionID, guildID)
		if ok {
			p.Destroy(c.Context())
			manager.DeletePlayer(sessionID, guildID)
		}
		return c.SendStatus(204)
	}
}

func mapPlayerToResponse(p *player.Player) PlayerResponse {
	return PlayerResponse{
		GuildID: p.GuildID,
		Track:   p.Track,
		Volume:  p.Volume,
		Paused:  p.Paused,
		State:   p.State(),
		Filters: p.Filters,
	}
}
