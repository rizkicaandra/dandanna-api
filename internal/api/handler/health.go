package handler

import (
	"net/http"

	"github.com/rizkicandra/dandanna-api/internal/api/dto"
	"github.com/rizkicandra/dandanna-api/internal/api/response"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/logger"
)

// Health handles health check requests
type Health struct {
	logger  logger.Logger
	version string
}

// NewHealth creates a new health handler
func NewHealth(log logger.Logger, version string) *Health {
	return &Health{logger: log, version: version}
}

// Handle handles GET /api/health.
// Method enforcement is done by the router (Go 1.22+ pattern matching), not here.
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
