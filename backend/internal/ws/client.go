package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

type Client struct {
	conn         *websocket.Conn
	room         *Room
	userID       string
	username     string
	playerNumber int
	send         chan ServerMessage
	done         chan struct{}
	once         sync.Once
}

func NewClient(conn *websocket.Conn, room *Room, userID, username string, playerNumber int) *Client {
	return &Client{
		conn:         conn,
		room:         room,
		userID:       userID,
		username:     username,
		playerNumber: playerNumber,
		send:         make(chan ServerMessage, 64),
		done:         make(chan struct{}),
	}
}

func (c *Client) ReadPump(ctx context.Context) {
	defer func() {
		c.room.unregister <- c
		c.Close()
	}()

	for {
		_, data, err := c.conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) != -1 {
				log.Printf("WebSocket closed for user %s: %v", c.userID, err)
			}
			return
		}

		var msg ClientMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			c.Send(ErrorMsg("invalid message format", "INVALID_FORMAT"))
			continue
		}

		switch msg.Type {
		case "ping":
			c.Send(PongMsg())
		case "sync_request":
			state := c.room.engine.State()
			c.Send(StateUpdateMsg(state))
		case "action":
			if msg.Data != nil {
				msg.Data.PlayerNumber = c.playerNumber
				c.room.actions <- msg.Data
			}
		}
	}
}

func (c *Client) WritePump(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err = c.conn.Write(ctx, websocket.MessageText, data)
			cancel()
			if err != nil {
				return
			}
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err := c.conn.Ping(ctx)
			cancel()
			if err != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}

func (c *Client) Send(msg ServerMessage) {
	select {
	case c.send <- msg:
	default:
		// Channel full, close client
		c.Close()
	}
}

func (c *Client) Close() {
	c.once.Do(func() {
		close(c.done)
		c.conn.Close(websocket.StatusNormalClosure, "")
	})
}
