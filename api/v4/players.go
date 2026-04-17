package v4

import (
	"github.com/disgoorg/snowflake/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/golink-devs/golink/internal/hub"
	"github.com/golink-devs/golink/internal/player"
	"github.com/golink-devs/golink/internal/sources"
	"github.com/golink-devs/golink/internal/voice"
)

type PlayerResponse struct {
	GuildID string         `json:"guildId"`
	Track   *sources.Track `json:"track"`
	Volume  int            `json:"volume"`
	Paused  bool           `json:"paused"`
	State   player.PlayerState `json:"state"`
	Voice   *VoiceState    `json:"voice"`
	Filters player.Filters `json:"filters"`
}

type VoiceState struct {
	Token     string `json:"token"`
	Endpoint  string `json:"endpoint"`
	SessionID string `json:"sessionId"`
}

type UpdatePlayerRequest struct {
	Track      *UpdateTrackRequest `json:"track,omitempty"`
	Identifier string              `json:"identifier,omitempty"`
	Position   *int64              `json:"position,omitempty"`
	EndTime    *int64              `json:"endTime,omitempty"`
	Volume     *int                `json:"volume,omitempty"`
	Paused     *bool               `json:"paused,omitempty"`
	Filters    *player.Filters     `json:"filters,omitempty"`
	Voice      *VoiceState         `json:"voice,omitempty"`
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
			return c.Status(404).SendString("Player not found")
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
			return c.Status(400).SendString(err.Error())
		}

		p, ok := manager.GetPlayer(sessionID, guildID)
		if !ok {
			client, ok := h.GetClient(sessionID)
			if !ok {
				return c.Status(400).SendString("Session not found")
			}
			p = manager.CreatePlayer(sessionID, client.UserID, guildID)
		}

		if noReplace && p.Track != nil {
			return c.JSON(mapPlayerToResponse(p))
		}

		// Handle track update
		var track *sources.Track
		if req.Track != nil && req.Track.Encoded != "" {
			info, err := sources.DecodeTrack(req.Track.Encoded)
			if err == nil {
				track = &sources.Track{
					Encoded: req.Track.Encoded,
					Info:    *info,
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
			// Resolve stream URL
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
				p.Play(c.Context(), *track, streamURL)
			}
		}

		if req.Volume != nil {
...
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
			guildID, _ := snowflake.Parse(p.GuildID)
			if p.VoiceConn == nil {
				p.VoiceConn = voice.NewVoiceConn(userID, guildID)
			}
			// Endpoint might need to be trimmed of :80 suffix
			// Lavalink clients usually provide it with :80 or :443
			// disgo handles this but we should be aware.

			// TODO: Find channelID from somewhere?
			// Actually Lavalink's PATCH doesn't provide channelID in VoiceState.
			// It only provides token, endpoint, sessionId.
			// Wait, the client usually joins the channel BEFORE sending this.
			// In Lavalink v4, the channelID is NOT in the voice state.
			// But disgo's HandleVoiceStateUpdate expects it.
			// Actually, if we don't have it, we might have issues.
			// Let's assume we can use a dummy for now or find it later.
			// In a real bot, you'd know which channel you are joining.
			// GoLink receives this from the client.

			// Let's check disgo voice.Conn.Open
			// err := p.VoiceConn.Open(c.Context(), channelID, userID, req.Voice.SessionID, req.Voice.Token, req.Voice.Endpoint)
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
	// TODO: Handle Voice field mapping if implemented
	return PlayerResponse{
		GuildID: p.GuildID,
		Track:   p.Track,
		Volume:  p.Volume,
		Paused:  p.Paused,
		State:   p.State(),
		Filters: p.Filters,
	}
}
