package ws

import (
	"context"
	"log/slog"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/peter/tacticarium/backend/internal/game"
)

var tracer = otel.Tracer("tacticarium/ws")

type Room struct {
	gameID     string
	engine     *game.Engine
	clients    map[*Client]bool
	actions    chan *game.GameAction
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex

	// Called when game state changes, for persistence
	OnStateChange func(state game.GameState, events []game.GameEvent)
}

func NewRoom(gameID string, engine *game.Engine) *Room {
	return &Room{
		gameID:     gameID,
		engine:     engine,
		clients:    make(map[*Client]bool),
		actions:    make(chan *game.GameAction, 64),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (r *Room) Run() {
	for {
		select {
		case client := <-r.register:
			// Snapshot peers BEFORE adding the new client, so we can replay
			// player_connected events to the joiner for everyone already here.
			r.mu.Lock()
			type peer struct {
				playerNumber int
				username     string
			}
			peers := make([]peer, 0, len(r.clients))
			for c := range r.clients {
				if c.isSpectator {
					continue
				}
				peers = append(peers, peer{c.playerNumber, c.username})
			}
			r.clients[client] = true
			r.mu.Unlock()

			if !client.isSpectator {
				// Ensure the player exists in the engine state.
				// This handles the case where player 2 joins after the room
				// was created by player 1's connection.
				r.engine.AddPlayer(&game.PlayerState{
					UserID:       client.userID,
					Username:     client.username,
					PlayerNumber: client.playerNumber,
					VPPaint:      game.MaxVPPaint,
				})

				// Notify others. Spectators are silent — players are not told
				// when one connects or leaves.
				r.broadcastExcept(PlayerConnectedMsg(client.playerNumber, client.username), client)
			}

			// Send current state to new client
			state := r.engine.State()
			client.Send(StateUpdateMsg(state))

			// Inform the joiner about peers that were already connected, so
			// the client knows the opponent is present without needing them
			// to act first.
			for _, p := range peers {
				client.Send(PlayerConnectedMsg(p.playerNumber, p.username))
			}

		case client := <-r.unregister:
			r.mu.Lock()
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
			}
			r.mu.Unlock()

			if !client.isSpectator {
				r.broadcast(PlayerDisconnectedMsg(client.playerNumber))
			}

		case action := <-r.actions:
			r.processAction(action)
		}
	}
}

func (r *Room) processAction(action *game.GameAction) {
	ctx, span := tracer.Start(context.Background(), "ws.processAction")
	span.SetAttributes(
		attribute.String("game.id", r.gameID),
		attribute.Int("game.player_number", action.PlayerNumber),
		attribute.String("game.action_type", string(action.Type)),
	)

	// Enrich span with user identity from the client that sent the action
	r.mu.RLock()
	for client := range r.clients {
		if client.playerNumber == action.PlayerNumber {
			span.SetAttributes(
				attribute.String("user.id", client.userID),
				attribute.String("user.name", client.username),
			)
			break
		}
	}
	r.mu.RUnlock()
	defer span.End()

	events, err := r.engine.Apply(ctx, *action)
	if err != nil {
		span.RecordError(err)
		// Find the client that sent this action and send error
		r.mu.RLock()
		for client := range r.clients {
			if client.playerNumber == action.PlayerNumber {
				client.Send(ErrorMsg(err.Error(), "ACTION_REJECTED"))
				break
			}
		}
		r.mu.RUnlock()
		return
	}

	// Persist first so each event has its assigned DB id (used by clients to
	// dedupe between live WS events and the REST history they fetch on load).
	state := r.engine.State()
	if r.OnStateChange != nil {
		r.OnStateChange(state, events)
	}

	// Broadcast events
	for _, event := range events {
		r.broadcast(EventMsg(event))
	}

	// Broadcast updated state
	r.broadcast(StateUpdateMsg(state))
}

func (r *Room) broadcast(msg ServerMessage) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for client := range r.clients {
		client.Send(msg)
	}
}

func (r *Room) broadcastExcept(msg ServerMessage, exclude *Client) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for client := range r.clients {
		if client != exclude {
			client.Send(msg)
		}
	}
}

func (r *Room) ClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}

func (r *Room) HasPlayer(userID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for client := range r.clients {
		if client.userID == userID {
			return true
		}
	}
	return false
}

func (r *Room) Register(client *Client) {
	r.register <- client
}

var _ = slog.Info // Ensure slog is used
