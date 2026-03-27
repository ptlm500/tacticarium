package ws

import "github.com/peter/tacticarium/backend/internal/game"

// Client -> Server messages
type ClientMessage struct {
	Type string          `json:"type"` // "action", "ping", "sync_request"
	Data *game.GameAction `json:"data,omitempty"`
}

// Server -> Client messages
type ServerMessage struct {
	Type string `json:"type"` // "state_update", "event", "error", "pong", "player_connected", "player_disconnected"
	Data any    `json:"data,omitempty"`
}

func StateUpdateMsg(state game.GameState) ServerMessage {
	return ServerMessage{Type: "state_update", Data: state}
}

func EventMsg(event game.GameEvent) ServerMessage {
	return ServerMessage{Type: "event", Data: event}
}

func ErrorMsg(message, code string) ServerMessage {
	return ServerMessage{Type: "error", Data: map[string]string{"message": message, "code": code}}
}

func PlayerConnectedMsg(playerNumber int, username string) ServerMessage {
	return ServerMessage{Type: "player_connected", Data: map[string]any{"playerNumber": playerNumber, "username": username}}
}

func PlayerDisconnectedMsg(playerNumber int) ServerMessage {
	return ServerMessage{Type: "player_disconnected", Data: map[string]any{"playerNumber": playerNumber}}
}

func PongMsg() ServerMessage {
	return ServerMessage{Type: "pong"}
}
