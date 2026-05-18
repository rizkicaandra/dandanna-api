package repository

import (
	"context"
	"time"

	"github.com/rizkicandra/dandanna-api/internal/domain/entity"
)

// TokenRepository manages opaque access and refresh tokens in Redis.
// All keys are namespaced by app (e.g. "artist") to prevent cross-app token misuse.
type TokenRepository interface {
	// Access tokens
	StoreAccess(ctx context.Context, app, token string, s *entity.Session, ttl time.Duration) error
	GetAccess(ctx context.Context, app, token string) (*entity.Session, error)
	DeleteAccess(ctx context.Context, app, token string) error

	// Refresh tokens
	StoreRefresh(ctx context.Context, app, token, userID string, ttl time.Duration) error
	GetRefresh(ctx context.Context, app, token string) (userID string, err error)
	DeleteRefresh(ctx context.Context, app, token string) error

	// Session pointer — enforces single active session per user per app.
	// Stores {refresh_token, access_token} so both can be revoked atomically.
	GetSession(ctx context.Context, app, userID string) (*entity.SessionPointer, error)
	SetSession(ctx context.Context, app, userID string, sp *entity.SessionPointer, ttl time.Duration) error
	DeleteSession(ctx context.Context, app, userID string) error

	// Login attempt tracking — IP-based brute-force protection.
	GetLoginAttempts(ctx context.Context, app, ip string) (int64, error)
	IncrLoginAttempts(ctx context.Context, app, ip string, window time.Duration) (int64, error)
	ResetLoginAttempts(ctx context.Context, app, ip string) error
}
