package hub

import (
	"encoding/json"
	"sync"

	"github.com/gofiber/websocket/v2"
	"github.com/rs/zerolog/log"
)

type Client struct {
	ID     string // sessionId
	UserID string // bot user ID
	Conn   *websocket.Conn
	Send   chan []byte
}

type Hub struct {
	clients    map[string]*Client // sessionId -> Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMsg
	mu         sync.RWMutex
}

type BroadcastMsg struct {
	GuildID   string // if set, only send to sessions with players for this guild
	SessionID string // if set, only send to this session
	Data      []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMsg),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			log.Info().Str("session_id", client.ID).Str("user_id", client.UserID).Msg("Client registered")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
				log.Info().Str("session_id", client.ID).Msg("Client unregistered")
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			if msg.SessionID != "" {
				if client, ok := h.clients[msg.SessionID]; ok {
					client.Send <- msg.Data
				}
			} else {
				// For now, broadcast to all.
				// Later we can filter by GuildID if needed.
				for _, client := range h.clients {
					client.Send <- msg.Data
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) Send(sessionID string, data interface{}) {
	payload, err := json.Marshal(data)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal message")
		return
	}
	h.broadcast <- &BroadcastMsg{
		SessionID: sessionID,
		Data:      payload,
	}
}

func (h *Hub) GetClient(sessionID string) (*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	client, ok := h.clients[sessionID]
	return client, ok
}
