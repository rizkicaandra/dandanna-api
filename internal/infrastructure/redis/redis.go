package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Config holds everything needed to connect to Redis.
// Populated from infrastructure/config — no direct coupling.
type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// pingTimeout is the maximum time allowed for the initial connectivity check.
const pingTimeout = 5 * time.Second

// Client wraps go-redis and implements the Pinger interface
// used by the health handler — keeping infrastructure decoupled from handlers.
type Client struct {
	rdb *goredis.Client
}

// New creates a Redis client and verifies connectivity with a ping.
// Returns an error if Redis is unreachable.
//
// The caller must call Close() when the application shuts down.
func New(ctx context.Context, cfg Config) (*Client, error) {
	rdb := goredis.NewClient(&goredis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("redis: ping failed (host=%s:%d db=%d): %w",
			cfg.Host, cfg.Port, cfg.DB, err)
	}

	return &Client{rdb: rdb}, nil
}

// PingContext satisfies the handler.Pinger interface.
// Used by the readyz endpoint to check Redis health.
func (c *Client) PingContext(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// Close closes the Redis connection pool.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Unwrap returns the underlying go-redis client for advanced operations
// (pub/sub, pipelines, etc.) in application services.
func (c *Client) Unwrap() *goredis.Client {
	return c.rdb
}
