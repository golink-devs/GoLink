package player

import (
	"sync"

	"github.com/golink-devs/golink/internal/hub"
)

type Manager struct {
	sessions map[string]map[string]*Player // sessionId -> guildId -> Player
	hub      *hub.Hub
	mu       sync.RWMutex
}

func NewManager(h *hub.Hub) *Manager {
	return &Manager{
		sessions: make(map[string]map[string]*Player),
		hub:      h,
	}
}

func (m *Manager) CreatePlayer(sessionID, userID, guildID string) *Player {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.sessions[sessionID]; !ok {
		m.sessions[sessionID] = make(map[string]*Player)
	}

	player := NewPlayer(sessionID, userID, guildID, m.hub)
	m.sessions[sessionID][guildID] = player
	return player
}

func (m *Manager) GetPlayer(sessionID, guildID string) (*Player, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, false
	}

	player, ok := session[guildID]
	return player, ok
}

func (m *Manager) GetPlayers(sessionID string) []*Player {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil
	}

	players := make([]*Player, 0, len(session))
	for _, p := range session {
		players = append(players, p)
	}
	return players
}

func (m *Manager) DeletePlayer(sessionID, guildID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, ok := m.sessions[sessionID]; ok {
		delete(session, guildID)
	}
}

func (m *Manager) DeleteSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, sessionID)
}

func (m *Manager) Sessions() map[string]map[string]*Player {
	m.mu.RLock()
	defer m.mu.RUnlock()

	copy := make(map[string]map[string]*Player)
	for k, v := range m.sessions {
		sessionCopy := make(map[string]*Player)
		for guildID, p := range v {
			sessionCopy[guildID] = p
		}
		copy[k] = sessionCopy
	}
	return copy
}
