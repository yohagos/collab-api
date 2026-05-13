package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yohagos/collab-api/internal/adapter/middleware"
	"github.com/yohagos/collab-api/internal/domain"
	"github.com/yohagos/collab-api/internal/infrastructure/config"
	"go.uber.org/zap"
)

type AuthHandler struct {
	config *config.Config
	logger *zap.Logger
}

func NewAuthHandler(cfg *config.Config, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{config: cfg, logger: logger}
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn int `json:"expires_in"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userID := domain.UserID{}
	usernane := req.Username

	accessClaim := middleware.Claims{
		UserID: userID,
		Username: usernane,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.config.JWT.AccessExpiry)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaim)
	accessStr, _ := accessToken.SignedString([]byte(h.config.JWT.Secret))

	refreshClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.config.JWT.RefreshExpiry)),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, _ := refreshToken.SignedString([]byte(h.config.JWT.Secret))

	resp := LoginResponse{
		AccessToken: accessStr,
		RefreshToken: refreshStr,
		ExpiresIn: int(h.config.JWT.AccessExpiry.Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	h.logger.Info("User logged in", zap.String("username", usernane))
}
