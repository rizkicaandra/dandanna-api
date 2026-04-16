package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib" // registers "pgx" driver
)

// Config holds everything needed to open a PostgreSQL connection pool.
// Populated from infrastructure/config — no direct coupling.
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// pingTimeout is the maximum time allowed for the initial connectivity check.
const pingTimeout = 5 * time.Second

// New opens a PostgreSQL connection pool, configures it, and verifies
// connectivity with a ping. Returns an error if the DB is unreachable.
//
// The caller must call db.Close() when the application shuts down.
func New(ctx context.Context, cfg Config) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: open: %w", err)
	}

	// Pool configuration — sized from config, not hardcoded
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Fail fast — if the DB is unreachable at startup, exit immediately
	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres: ping failed (host=%s:%d db=%s): %w",
			cfg.Host, cfg.Port, cfg.Database, err)
	}

	return db, nil
}
