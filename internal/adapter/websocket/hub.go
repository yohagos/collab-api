package websocket

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/yohagos/collab-api/internal/domain"
	"go.uber.org/zap"
)

type Hub struct {
	Rooms      map[domain.RoomID]map[string]*Client
	Broadcast  chan RoomMessage
	Register   chan *Client
	Unregister chan *Client
	Logger     *zap.Logger
	mu         sync.RWMutex
}

type RoomMessage struct {
	RoomID  domain.RoomID   `json:"room_id"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		Rooms:      make(map[domain.RoomID]map[string]*Client),
		Broadcast:  make(chan RoomMessage, 512),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Logger:     logger,
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.Register:
			h.mu.Lock()
			if _, exists := h.Rooms[client.RoomID]; !exists {
				h.Rooms[client.RoomID] = make(map[string]*Client)
			}
			h.Rooms[client.RoomID][client.ID] = client
			h.mu.Unlock()

			h.Logger.Info(
				"client registered",
				zap.String("client_id", client.ID),
				zap.String("room_id", client.RoomID.String()),
			)

		case client := <-h.Unregister:
			h.mu.Lock()
			if clients, ok := h.Rooms[client.RoomID]; ok {
				if _, exists := clients[client.ID]; exists {
					delete(clients, client.ID)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.Rooms, client.RoomID)
					}
				}
			}
			h.mu.Unlock()

		case msg := <-h.Broadcast:
			h.mu.RLock()
			if clients, ok := h.Rooms[msg.RoomID]; ok {
				for _, client := range clients {
					select {
					case client.Send <- append([]byte(`{"type":"` + msg.Type + `", "payload":`), append(msg.Payload, '}')...):
					default:
						go func(c *Client) {
							h.Unregister <- c
						}(client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastToRoom(roomID domain.RoomID, msgType string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		h.Logger.Error("failed to marshal broadcast payload", zap.Error(err))
	}

	h.Broadcast <- RoomMessage{
		RoomID:  roomID,
		Type:    msgType,
		Payload: data,
	}
}
