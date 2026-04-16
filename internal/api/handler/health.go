package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/rizkicandra/dandanna-api/internal/api/dto"
	"github.com/rizkicandra/dandanna-api/internal/api/response"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/logger"
)

// Pinger is satisfied by *sqlx.DB and *redis.Client — keeps the handler
// decoupled from concrete infrastructure types and easy to test.
type Pinger interface {
	PingContext(ctx context.Context) error
}

// Health handles liveness and readiness check requests
type Health struct {
	logger  logger.Logger
	version string
	db      Pinger
	redis   Pinger
}

// NewHealth creates a new health handler
func NewHealth(log logger.Logger, version string, db Pinger, redis Pinger) *Health {
	return &Health{logger: log, version: version, db: db, redis: redis}
}

// Handle handles GET /api/health — liveness check.
// Always returns 200 if the process is running.
func (h *Health) Handle(w http.ResponseWriter, r *http.Request) {
	data := dto.HealthData{
		Status:  "healthy",
		Version: h.version,
	}

	if err := response.JSON(w, r, data, http.StatusOK); err != nil {
		h.logger.Error("failed to write health response", logger.Err(err))
	}

	h.logger.Debug("health check", logger.String("ip", r.RemoteAddr))
}

// Readyz handles GET /api/readyz — readiness check.
// Returns 200 only when all dependencies are reachable.
// Kubernetes uses this to decide whether to send traffic to the pod.
func (h *Health) Readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	var errors []dto.ErrorItem

	if err := h.db.PingContext(ctx); err != nil {
		h.logger.Warn("readyz: postgres unreachable", logger.Err(err))
		errors = append(errors, response.NewError("", "DB_UNAVAILABLE", "database is not reachable"))
	}

	if err := h.redis.PingContext(ctx); err != nil {
		h.logger.Warn("readyz: redis unreachable", logger.Err(err))
		errors = append(errors, response.NewError("", "REDIS_UNAVAILABLE", "redis is not reachable"))
	}

	if len(errors) > 0 {
		response.Error(w, r, http.StatusServiceUnavailable, errors...)
		return
	}

	response.JSON(w, r, dto.ReadyzData{
		Postgres: "ok",
		Redis:    "ok",
	}, http.StatusOK)
}
