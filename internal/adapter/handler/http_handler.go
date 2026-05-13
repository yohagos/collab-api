package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yohagos/collab-api/internal/adapter/middleware"
	"github.com/yohagos/collab-api/internal/application"
	"github.com/yohagos/collab-api/internal/domain"
	"go.uber.org/zap"
)

type HTTPHandler struct {
	roomService     *application.RoomService
	messageService  *application.MessageService
	presenceService *application.PresenceService
	logger          *zap.Logger
}

func NewHTTPHandler(
	roomService *application.RoomService,
	messageService *application.MessageService,
	presenceService *application.PresenceService,
	logger *zap.Logger,
) *HTTPHandler {
	return &HTTPHandler{
		roomService:     roomService,
		messageService:  messageService,
		presenceService: presenceService,
		logger:          logger,
	}
}

func (h *HTTPHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name" validate:",min=3,max=100"`
		Description string `json:"description,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not found in context")
		return
	}

	room, err := h.roomService.CreateRoom(r.Context(), req.Name, req.Description, claims.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(room)
}

func (h *HTTPHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	rooms, err := h.roomService.ListRooms(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch rooms")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rooms": rooms,
		"count": len(rooms),
	})
}

func (h *HTTPHandler) GetRoomMembers(w http.ResponseWriter, r *http.Request) {
	roomIDStr := mux.Vars(r)["room_id"]
	if roomIDStr == "" {
		writeError(w, http.StatusBadRequest, "room_id is required")
		return
	}
	roomID := domain.RoomID(uuid.MustParse(roomIDStr))

	members, err := h.presenceService.GetOnlineUsers(r.Context(), roomID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get online users")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"room_id": roomID,
		"count":   len(members),
		"members": members,
	})
}

func (h *HTTPHandler) GetRoom(w http.ResponseWriter, r *http.Request) {
	roomdIDStr := mux.Vars(r)["room_id"]
	if roomdIDStr == "" {
		writeError(w, http.StatusBadRequest, "room_id is required")
		return
	}

	roomID := domain.RoomID(uuid.MustParse(roomdIDStr))
	room, err := h.roomService.GetRoom(r.Context(), roomID)
	if err != nil {
		if err == domain.ErrRoomNotFound {
			writeError(w, http.StatusNotFound, "room not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch room")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(room)
}

func (h *HTTPHandler) GetRoomMessages(w http.ResponseWriter, r *http.Request) {
	roomIDStr := mux.Vars(r)["room_id"]
	if roomIDStr == "" {
		writeError(w, http.StatusBadRequest, "room_id is required")
		return
	}

	roomID := domain.RoomID(uuid.MustParse(roomIDStr))
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	messages, err := h.messageService.GetMessagesByRoom(r.Context(), roomID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch messages")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"room_id":  roomID,
		"count":    len(messages),
		"messages": messages,
	})
}

func writeError(w http.ResponseWriter, status uint, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(status))
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
