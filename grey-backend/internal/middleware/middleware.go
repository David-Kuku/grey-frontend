package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/David-Kuku/grey-frontend/grey-backend/internal/auth"
	"github.com/David-Kuku/grey-frontend/grey-backend/internal/models"
	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	RequestIDKey contextKey = "request_id"
)
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
		w.Header().Set("X-Request-ID", reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(ww, r)

			reqID, _ := r.Context().Value(RequestIDKey).(string)
			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.statusCode,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", reqID,
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
func Auth(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				writeError(w, http.StatusUnauthorized, "AUTH_REQUIRED", "Authorization header is required")
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				writeError(w, http.StatusUnauthorized, "INVALID_AUTH_FORMAT", "Authorization header must be: Bearer <token>")
				return
			}

			claims, err := authService.ValidateToken(parts[1])
			if err != nil {
				writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Token is invalid or expired")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
func GetUserID(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(UserIDKey).(uuid.UUID)
	return id
}
func GetRequestID(ctx context.Context) string {
	id, _ := ctx.Value(RequestIDKey).(string)
	return id
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := models.APIError{Code: code, Message: message}
	w.Write([]byte(`{"code":"` + resp.Code + `","message":"` + resp.Message + `"}`))
}
