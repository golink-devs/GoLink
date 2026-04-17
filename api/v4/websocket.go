package v4

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gofiber/websocket/v2"
	"github.com/golink-devs/golink/api/v1"
	"github.com/golink-devs/golink/internal/hub"
	"github.com/rs/zerolog/log"
)

func NewWebSocketHandler(h *hub.Hub) func(*websocket.Conn) {
	return func(c *websocket.Conn) {
		v1.WebSocketConnections.Inc()
		defer v1.WebSocketConnections.Dec()

		userID := c.Query("User-Id")
		if userID == "" {
			userID = c.Headers("User-Id")
		}

		sessionID := c.Query("Session-Id")
		if sessionID == "" {
			sessionID = c.Headers("Session-Id")
		}

		if sessionID == "" {
			b := make([]byte, 16)
			rand.Read(b)
			sessionID = hex.EncodeToString(b)
		}

		client := &hub.Client{
			ID:     sessionID,
			UserID: userID,
			Conn:   c,
			Send:   make(chan []byte, 256),
		}

		h.Register(client)

		// Send ready op
		h.Send(sessionID, map[string]interface{}{
			"op":        "ready",
			"resumed":   false,
			"sessionId": sessionID,
		})

		// Read loop (clients don't usually send much, but we must keep it open)
		go func() {
			for {
				_, _, err := c.ReadMessage()
				if err != nil {
					log.Debug().Err(err).Str("session_id", sessionID).Msg("WebSocket read error")
					break
				}
			}
			h.Unregister(client)
		}()

		// Write loop
		for msg := range client.Send {
			if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Debug().Err(err).Str("session_id", sessionID).Msg("WebSocket write error")
				break
			}
		}
		c.Close()
	}
}
