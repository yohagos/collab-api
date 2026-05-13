package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RateLimiter struct {
	rds    *redis.Client
	logger *zap.Logger
	mu     sync.RWMutex
}

func NewRateLimiter(redisClient *redis.Client, logger *zap.Logger) *RateLimiter {
	return &RateLimiter{
		rds:    redisClient,
		logger: logger,
	}
}

func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		key := "ratelimit:" + ip

		count, err := rl.rds.Incr(r.Context(), key).Result()
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		if count == 1 {
			rl.rds.Expire(r.Context(), key, 60*time.Second)
		}

		if count > 30 {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			rl.logger.Warn("Rate Limit exceeded", zap.String("ip", ip))
			return
		}

		next.ServeHTTP(w, r)
	})
}
