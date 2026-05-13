package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type HealthHandler struct {
	db *pgxpool.Pool
	rds *redis.Client
	logger *zap.Logger
}

func NewHealthHandler(
	db *pgxpool.Pool,
	rdsClient *redis.Client,
	lgr *zap.Logger,
) *HealthHandler {
	return &HealthHandler{
		db: db,
		rds: rdsClient,
		logger: lgr,
	}
}

type HealthResponse struct {
	Status string `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Services map[string]string `json:"services"`
}

func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	services := make(map[string]string)

	pgCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.db.Ping(pgCtx); err == nil {
		services["postgres"] = "healthy"
	} else {
		services["postgres"] = "unhealthy"
	}

	redisCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.rds.Ping(redisCtx).Err(); err == nil {
		services["redis"] = "healthy"
	} else {
		services["redis"] = "unhealthy"
	}

	status := "healthy"
	for _, s := range services {
		if s == "unhealthy" {
			status = "degraded"
			break
		}
	}

	resp := HealthResponse{
		Status: status,
		Timestamp: time.Now().UTC(),
		Services: services,
	}

	w.Header().Set("Content-Type", "application/json")
	if status == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusPartialContent)
	}
	json.NewEncoder(w).Encode(resp)
}