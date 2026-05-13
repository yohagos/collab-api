package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/yohagos/collab-api/internal/adapter/middleware"
	cws "github.com/yohagos/collab-api/internal/adapter/websocket"
	"github.com/yohagos/collab-api/internal/application"
	"github.com/yohagos/collab-api/internal/domain"
	"github.com/yohagos/collab-api/internal/infrastructure/config"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WSHandler struct {
	hub             *cws.Hub
	roomService     *application.RoomService
	messageService  *application.MessageService
	presenceService *application.PresenceService
	config          *config.Config
	logger          *zap.Logger
}

func NewWSHandler(
	hub *cws.Hub,
	roomService *application.RoomService,
	messageService *application.MessageService,
	presenceService *application.PresenceService,
	cfg *config.Config,
	logger *zap.Logger,
) *WSHandler {
	return &WSHandler{
		hub:             hub,
		roomService:     roomService,
		messageService:  messageService,
		presenceService: presenceService,
		config:          cfg,
		logger:          logger,
	}
}

func (h *WSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "token" {
		token = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		token = strings.TrimSpace(token)
	}

	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	claims := &middleware.Claims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(h.config.JWT.Secret), nil
	})
	if err != nil || !parsedToken.Valid {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	roomIDStr := r.URL.Query().Get("room_id")
	if roomIDStr == "" {
		http.Error(w, "missing room_id", http.StatusBadRequest)
		return
	}

	roomID := domain.RoomID(uuid.MustParse(roomIDStr))

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", zap.Error(err))
		return
	}

	client := cws.NewClient(conn, claims.UserID, roomID, h.logger)

	h.hub.Register <- client

	_ = h.presenceService.UserJoined(r.Context(), roomID, claims.UserID, claims.Username)
	_ = h.roomService.JoinRoom(r.Context(), roomID, claims.UserID, claims.Username)

	go client.WritePump(r.Context())
	go h.readPump(client)

	h.logger.Info(
		"Client connected via WebSocket",
		zap.String("room_id", roomID.String()),
		zap.String("user_id", claims.UserID.String()),
		zap.String("username", claims.Username),
	)
}

func (h *WSHandler) readPump(client *cws.Client) {
	defer func() {
		h.hub.Unregister <- client
		h.presenceService.UserLeft(context.Background(), client.RoomID, client.UserID)
		client.Conn.Close()
	}()

	for {
		_, rawMessage, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}

		var msg map[string]interface{}
		if json.Unmarshal(rawMessage, &msg) != nil {
			content := string(rawMessage)
			if content != "" {
				savedMsg, _ := h.messageService.SaveMessage(context.Background(), client.RoomID, client.UserID, "USER", content)
				h.hub.BroadcastToRoom(client.RoomID, "chat", savedMsg)
			}
			continue
		}

		msgType, _ := msg["type"].(string)
		switch msgType {
		case "typing":
			_ = h.presenceService.SetUserTyping(context.Background(), client.RoomID, client.UserID, "USER")
			h.hub.BroadcastToRoom(client.RoomID, "typing", map[string]interface{}{
				"user_id":   client.UserID,
				"username":  "USER",
				"is_typing": true,
			})

		case "cursor":
			h.hub.BroadcastToRoom(client.RoomID, "cursor", map[string]interface{}{
				"user_id":  client.UserID,
				"username": "USER",
				"x":        msg["x"],
				"y":        msg["y"],
				"color":    msg["color"],
			})
		default:
			if content, ok := msg["content"]; ok && content != "" {
				savedMsg, _ := h.messageService.SaveMessage(
					context.Background(),
					client.RoomID,
					client.UserID,
					"USER",
					content.(string),
				)
				h.hub.BroadcastToRoom(client.RoomID, "chat", savedMsg)
			}
		}
	}
}
