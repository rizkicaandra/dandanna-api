package bootstrap

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/rizkicandra/dandanna-api/internal/api/handler"
	"github.com/rizkicandra/dandanna-api/internal/api/router"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/config"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/logger"
	redisinfra "github.com/rizkicandra/dandanna-api/internal/infrastructure/redis"
)

// NewRouter wires all handlers and returns a fully configured router.
func NewRouter(
	ctx context.Context,
	db *sqlx.DB,
	rdb *redisinfra.Client,
	cfg *config.Config,
	log logger.Logger,
	version string,
	corsOrigins []string,
) (*router.Router, error) {
	artistHandler, err := NewArtistHandler(ctx, db, rdb, cfg, log)
	if err != nil {
		return nil, err
	}

	health := handler.NewHealth(log, version, db, rdb)

	r := router.New(log, corsOrigins)
	r.Setup(health, artistHandler)
	return r, nil
}
