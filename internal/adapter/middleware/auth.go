package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yohagos/collab-api/internal/domain"
	"github.com/yohagos/collab-api/internal/infrastructure/config"
	"go.uber.org/zap"
)

type contextKey string

const UserContextKey contextKey = "user"

type Claims struct {
	UserID   domain.UserID `json:"user_id"`
	Username string        `json:"username"`
	jwt.RegisteredClaims
}

type AuthMiddleware struct {
	config *config.Config
	logger *zap.Logger
}

func NewAuthMiddleware(cfg *config.Config, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		config: cfg,
		logger: logger,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "invalid token format", http.StatusUnauthorized)
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.config.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*Claims)
	return claims, ok
}
