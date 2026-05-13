package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/yohagos/collab-api/internal/adapter/handler"
	"github.com/yohagos/collab-api/internal/infrastructure/config"
	"github.com/yohagos/collab-api/internal/infrastructure/database"
	"github.com/yohagos/collab-api/internal/infrastructure/logger"
	"github.com/yohagos/collab-api/internal/infrastructure/redis"
)

func main() {
	fmt.Println("Init main file")
	cfg := config.Load()

	fmt.Println("\nLoaded configs")

	log := logger.New(cfg)

	fmt.Println("\nContent of configs => ", cfg)

	db, err := database.NewPostgresConnection(cfg)
	if err != nil {
		log.Fatal("Failed to connect to Postgres database", zap.Error(err))
	}
	defer db.Close()

	redisClient := redis.NewClient(cfg)

	router := handler.NewRouter(cfg, log, db, redisClient)

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("Server starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server exited gracefully")
}
