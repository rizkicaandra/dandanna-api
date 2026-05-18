package bootstrap

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/rizkicandra/dandanna-api/internal/api/handler"
	"github.com/rizkicandra/dandanna-api/internal/api/router"
	redisinfra "github.com/rizkicandra/dandanna-api/internal/infrastructure/redis"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/logger"
)

// NewRouter wires all handlers and returns a fully configured router.
func NewRouter(
	ctx context.Context,
	db *sqlx.DB,
	rdb *redisinfra.Client,
	log logger.Logger,
	version string,
	corsOrigins []string,
) (*router.Router, error) {
	artistHandler, err := NewArtistHandler(ctx, db, log)
	if err != nil {
		return nil, err
	}

	health := handler.NewHealth(log, version, db, rdb)

	r := router.New(log, corsOrigins)
	r.Setup(health, artistHandler)
	return r, nil
}
