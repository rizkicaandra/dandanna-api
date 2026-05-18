package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rizkicandra/dandanna-api/internal/bootstrap"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/config"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/logger"
	pginfra "github.com/rizkicandra/dandanna-api/internal/infrastructure/postgres"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/redis"
)

// Version and BuildTime are injected at build time via ldflags:
// -X main.Version=$(git describe --tags) -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// ── 1. CLI flags ─────────────────────────────────────────────────────────
	// Supports: ./dandanna-api --version
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("dandanna-api version=%s built=%s\n", Version, BuildTime)
		os.Exit(0)
	}

	// ── 2. Configuration ──────────────────────────────────────────────────────
	// Loads .env for local development; ignored in production where the platform
	// (Docker / Kubernetes) injects real environment variables directly.
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// ── 3. Logger ─────────────────────────────────────────────────────────────
	// Structured JSON logger backed by zerolog. Log level comes from APP_LOG_LEVEL.
	log := logger.New(logger.LogLevel(cfg.App.LogLevel), nil)
	log.Info("starting dandanna-api",
		logger.String("version", Version),
		logger.String("build_time", BuildTime),
		logger.String("environment", cfg.App.Environment),
		logger.Int("port", cfg.Server.Port),
	)

	// ── 4. PostgreSQL ─────────────────────────────────────────────────────────
	// Primary data store. Fails fast on startup if unreachable — no point
	// serving traffic without a database.
	db, err := pginfra.New(context.Background(), pginfra.Config{
		Host:            cfg.Postgres.Host,
		Port:            cfg.Postgres.Port,
		User:            cfg.Postgres.User,
		Password:        cfg.Postgres.Password,
		Database:        cfg.Postgres.Database,
		SSLMode:         cfg.Postgres.SSLMode,
		MaxOpenConns:    cfg.Postgres.MaxOpenConns,
		MaxIdleConns:    cfg.Postgres.MaxIdleConns,
		ConnMaxLifetime: cfg.Postgres.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Postgres.ConnMaxIdleTime,
	})
	if err != nil {
		log.Error("failed to connect to postgres", logger.Err(err))
		os.Exit(1)
	}
	defer db.Close()
	log.Info("connected to postgres",
		logger.String("host", cfg.Postgres.Host),
		logger.Int("port", cfg.Postgres.Port),
		logger.String("database", cfg.Postgres.Database),
	)

	// ── 5. Redis ──────────────────────────────────────────────────────────────
	// Used for caching and session storage. Also fails fast on startup.
	rdb, err := redis.New(context.Background(), redis.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		log.Error("failed to connect to redis", logger.Err(err))
		os.Exit(1)
	}
	defer rdb.Close()
	log.Info("connected to redis",
		logger.String("host", cfg.Redis.Host),
		logger.Int("port", cfg.Redis.Port),
	)

	// ── 6. Bootstrap ──────────────────────────────────────────────────────────
	// Wires all features (handlers, services, repositories) and registers routes.
	// Each feature lives in its own bootstrap file under internal/bootstrap/.
	r, err := bootstrap.NewRouter(context.Background(), db, rdb, cfg, log, Version, cfg.App.CORSOrigins)
	if err != nil {
		log.Error("failed to bootstrap application", logger.Err(err))
		os.Exit(1)
	}

	// ── 7. HTTP server ────────────────────────────────────────────────────────
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r.Handler(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start the server in a goroutine so the main goroutine can listen for
	// shutdown signals below.
	serverErr := make(chan error, 1)
	go func() {
		log.Info("server listening", logger.String("addr", server.Addr))
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// ── 8. Graceful shutdown ──────────────────────────────────────────────────
	// Block until either the server crashes or we receive SIGINT / SIGTERM
	// (e.g. Ctrl+C or a Kubernetes pod termination signal).
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Error("server failed to start", logger.Err(err))
		os.Exit(1)
	case sig := <-quit:
		log.Info("shutdown signal received", logger.String("signal", sig.String()))
	}

	// Give in-flight requests up to ShutdownTimeout to finish before the
	// process exits. New requests are rejected immediately.
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", logger.Err(err))
		os.Exit(1)
	}

	log.Info("server stopped")
}
