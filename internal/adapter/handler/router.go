package handler

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/yohagos/collab-api/internal/adapter/middleware"
	"github.com/yohagos/collab-api/internal/adapter/repository"
	"github.com/yohagos/collab-api/internal/adapter/websocket"
	"github.com/yohagos/collab-api/internal/application"
	"github.com/yohagos/collab-api/internal/infrastructure/config"
	"go.uber.org/zap"
)

func NewRouter(
	cfg *config.Config,
	logger *zap.Logger,
	db *pgxpool.Pool,
	redisClient *redis.Client,
) http.Handler {
	roomRepo := repository.NewRoomRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	roomService := application.NewRoomService(roomRepo, logger)
	messageService := application.NewMessageService(messageRepo, logger)
	presenceService := application.NewPresenceService(redisClient, logger)

	hub := websocket.NewHub(logger)
	go hub.Run(context.Background())

	httpHandler := NewHTTPHandler(roomService, messageService, presenceService, logger)
	wsHandler := NewWSHandler(hub, roomService, messageService, presenceService, cfg, logger)
	healthHandler := NewHealthHandler(db, redisClient, logger)
	authHandler := NewAuthHandler(cfg, logger)
	authMiddleware := middleware.NewAuthMiddleware(cfg, logger)

	rateLimiter := middleware.NewRateLimiter(redisClient, logger)

	router := mux.NewRouter()

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}).Methods("GET")

	router.HandleFunc("/healthz", healthHandler.HealthCheck).Methods("GET")

	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(rateLimiter.RateLimit)

	protected := api.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.Authenticate)

	protected.HandleFunc("/rooms", httpHandler.CreateRoom).Methods("POST")
	protected.HandleFunc("/rooms", httpHandler.ListRooms).Methods("GET")
	protected.HandleFunc("/rooms/{room_id}", httpHandler.GetRoom).Methods("GET")
	protected.HandleFunc("/rooms/{room_id}/members", httpHandler.GetRoomMembers).Methods("GET")
	protected.HandleFunc("/rooms/{room_id}/messages", httpHandler.GetRoomMessages).Methods("GET")

	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST")

	router.HandleFunc("/ws", wsHandler.ServeWS).Methods("GET")

	return router
}
