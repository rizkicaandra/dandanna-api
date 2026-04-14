package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/rizkicandra/dandanna-api/internal/infrastructure/logger"
)

type contextKey string

const requestIDKey contextKey = "requestId"

// RequestIDFromContext extracts the request ID from the context.
// Returns an empty string if not present.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// RequestID reads X-Request-ID from the incoming request (or generates one),
// sets it on the response header, and injects it into the request context.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		w.Header().Set("X-Request-ID", requestID)

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logging logs each HTTP request with method, path, status, duration, and request ID.
func Logging(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)

			log.Info("request completed",
				logger.String("method", r.Method),
				logger.String("path", r.URL.Path),
				logger.Int("status", rw.status),
				logger.Duration("duration", time.Since(start)),
				logger.String("requestId", RequestIDFromContext(r.Context())),
				logger.String("ip", r.RemoteAddr),
			)
		})
	}
}

// Recovery recovers from panics, logs them with the request ID, and returns 500.
func Recovery(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered",
						logger.Any("error", rec),
						logger.String("path", r.URL.Path),
						logger.String("requestId", RequestIDFromContext(r.Context())),
					)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// CORS sets CORS headers based on a whitelist of allowed origins.
// Pass allowedOrigins from config — do NOT use ["*"] in production.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[strings.ToLower(o)] = struct{}{}
	}
	wildcard := func() bool {
		_, ok := originSet["*"]
		return ok
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if wildcard {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" {
				if _, allowed := originSet[strings.ToLower(origin)]; allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders sets standard security response headers.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture the written status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// generateRequestID returns a cryptographically random 16-byte hex string.
func generateRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Extremely unlikely, but fall back to timestamp to avoid empty IDs
		return time.Now().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(b)
}
