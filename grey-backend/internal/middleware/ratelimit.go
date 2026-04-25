package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)
type RateLimiterConfig struct {
	Rate rate.Limit
	Burst int
	CleanupInterval time.Duration
	MaxIdleTime time.Duration
}

type userLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}
type PerUserRateLimiter struct {
	mu       sync.Mutex
	limiters map[uuid.UUID]*userLimiter
	cfg      RateLimiterConfig
}
func NewPerUserRateLimiter(cfg RateLimiterConfig) *PerUserRateLimiter {
	if cfg.CleanupInterval == 0 {
		cfg.CleanupInterval = 1 * time.Minute
	}
	if cfg.MaxIdleTime == 0 {
		cfg.MaxIdleTime = 5 * time.Minute
	}

	rl := &PerUserRateLimiter{
		limiters: make(map[uuid.UUID]*userLimiter),
		cfg:      cfg,
	}

	go rl.cleanup()
	return rl
}
func (rl *PerUserRateLimiter) getLimiter(userID uuid.UUID) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.limiters[userID]
	if !exists {
		entry = &userLimiter{
			limiter: rate.NewLimiter(rl.cfg.Rate, rl.cfg.Burst),
		}
		rl.limiters[userID] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter
}
func (rl *PerUserRateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cfg.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.cfg.MaxIdleTime)
		for id, entry := range rl.limiters {
			if entry.lastSeen.Before(cutoff) {
				delete(rl.limiters, id)
			}
		}
		rl.mu.Unlock()
	}
}
func (rl *PerUserRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r.Context())
		if userID == uuid.Nil {
			next.ServeHTTP(w, r)
			return
		}

		limiter := rl.getLimiter(userID)
		if !limiter.Allow() {
			w.Header().Set("Retry-After", "1")
			writeError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED",
				"Too many requests. Please slow down and try again shortly.")
			return
		}

		next.ServeHTTP(w, r)
	})
}
