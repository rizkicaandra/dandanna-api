package bootstrap

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/rizkicandra/dandanna-api/internal/api/handler"
	"github.com/rizkicandra/dandanna-api/internal/application/service"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/logger"
	pginfra "github.com/rizkicandra/dandanna-api/internal/infrastructure/postgres"
)

// NewArtistHandler resolves all dependencies for the artist feature and returns
// a ready-to-use handler. Fails fast if seed data is missing.
func NewArtistHandler(ctx context.Context, db *sqlx.DB, log logger.Logger) (*handler.Artist, error) {
	appID, err := pginfra.ResolveApplicationID(ctx, db, "ARTIST_PORTAL")
	if err != nil {
		return nil, err
	}

	roleID, err := pginfra.ResolveRoleID(ctx, db, "ARTIST")
	if err != nil {
		return nil, err
	}

	log.Info("resolved artist startup IDs",
		logger.Int64("app_id", appID),
		logger.Int64("role_id", roleID),
	)

	repo := pginfra.NewArtistRepository(db)
	svc := service.NewArtistService(repo, log, appID, roleID)
	return handler.NewArtist(log, svc), nil
}
