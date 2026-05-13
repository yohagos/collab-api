package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/yohagos/collab-api/internal/domain"
	"go.uber.org/zap"
)

type Client struct {
	ID     string
	UserID domain.UserID
	RoomID domain.RoomID
	Conn   *websocket.Conn
	Send   chan []byte
	Logger *zap.Logger
	mu     sync.Mutex
}

func NewClient(
	conn *websocket.Conn,
	userID domain.UserID,
	roomID domain.RoomID,
	logger *zap.Logger,
) *Client {
	return &Client{
		ID:     uuid.New().String(),
		UserID: userID,
		RoomID: roomID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Logger: logger,
	}
}

func (c *Client) WritePump(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-c.Send:
			c.mu.Lock()
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.mu.Unlock()
				return
			}

			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			c.mu.Unlock()
			if err != nil {
				c.Logger.Error("write error", zap.Error(err))
				return
			}
		case <-ticker.C:
			c.mu.Lock()
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()
		}
	}
}
