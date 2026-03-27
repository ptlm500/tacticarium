package ws

import (
	"sync"

	"github.com/peter/tacticarium/backend/internal/game"
)

type Hub struct {
	rooms map[string]*Room
	mu    sync.RWMutex

	// Callback for persisting state changes
	OnStateChange func(state game.GameState, events []game.GameEvent)
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]*Room),
	}
}

func (h *Hub) GetOrCreateRoom(gameID string, engine *game.Engine) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.rooms[gameID]; ok {
		return room
	}

	room := NewRoom(gameID, engine)
	room.OnStateChange = h.OnStateChange
	h.rooms[gameID] = room
	go room.Run()

	return room
}

func (h *Hub) GetRoom(gameID string) *Room {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.rooms[gameID]
}

func (h *Hub) RemoveRoom(gameID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.rooms, gameID)
}
