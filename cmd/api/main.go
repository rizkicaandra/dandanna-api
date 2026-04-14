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

	"github.com/rizkicandra/dandanna-api/internal/api/handler"
	"github.com/rizkicandra/dandanna-api/internal/api/router"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/config"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/logger"
)

// Version and BuildTime are injected at build time via ldflags:
// -X main.Version=$(git describe --tags) -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("dandanna-api version=%s built=%s\n", Version, BuildTime)
		os.Exit(0)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(logger.LogLevel(cfg.App.LogLevel), nil)
	log.Info("starting dandanna-api",
		logger.String("version", Version),
		logger.String("build_time", BuildTime),
		logger.String("environment", cfg.App.Environment),
		logger.Int("port", cfg.Server.Port),
	)

	health := handler.NewHealth(log, Version)

	r := router.New(log, cfg.App.CORSOrigins)
	r.Setup(health)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r.Handler(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info("server listening", logger.String("addr", server.Addr))
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Error("server failed to start", logger.Err(err))
		os.Exit(1)
	case sig := <-quit:
		log.Info("shutdown signal received", logger.String("signal", sig.String()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", logger.Err(err))
		os.Exit(1)
	}

	log.Info("server stopped")
}
