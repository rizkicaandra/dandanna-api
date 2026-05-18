package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/rizkicandra/dandanna-api/internal/domain/entity"
	domainerror "github.com/rizkicandra/dandanna-api/internal/domain/error"
)

// TokenRepository implements domain/repository.TokenRepository using Redis.
// All keys are namespaced as <app>:<type>:<id> to prevent cross-app token misuse.
type TokenRepository struct {
	rdb *goredis.Client
}

// NewTokenRepository creates a TokenRepository backed by the given Redis client.
func NewTokenRepository(c *Client) *TokenRepository {
	return &TokenRepository{rdb: c.Unwrap()}
}

func (r *TokenRepository) StoreAccess(ctx context.Context, app, token string, s *entity.Session, ttl time.Duration) error {
	b, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("token: marshal session: %w", err)
	}
	return r.rdb.Set(ctx, accessKey(app, token), b, ttl).Err()
}

func (r *TokenRepository) GetAccess(ctx context.Context, app, token string) (*entity.Session, error) {
	raw, err := r.rdb.Get(ctx, accessKey(app, token)).Result()
	if errors.Is(err, goredis.Nil) {
		return nil, &domainerror.Unauthorized{Message: "invalid or expired access token"}
	}
	if err != nil {
		return nil, fmt.Errorf("token: get access: %w", err)
	}
	var s entity.Session
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return nil, fmt.Errorf("token: unmarshal session: %w", err)
	}
	return &s, nil
}

func (r *TokenRepository) DeleteAccess(ctx context.Context, app, token string) error {
	return r.rdb.Del(ctx, accessKey(app, token)).Err()
}

func (r *TokenRepository) StoreRefresh(ctx context.Context, app, token, userID string, ttl time.Duration) error {
	return r.rdb.Set(ctx, refreshKey(app, token), userID, ttl).Err()
}

func (r *TokenRepository) GetRefresh(ctx context.Context, app, token string) (string, error) {
	userID, err := r.rdb.Get(ctx, refreshKey(app, token)).Result()
	if errors.Is(err, goredis.Nil) {
		return "", &domainerror.Unauthorized{Message: "invalid or expired refresh token"}
	}
	return userID, err
}

func (r *TokenRepository) DeleteRefresh(ctx context.Context, app, token string) error {
	return r.rdb.Del(ctx, refreshKey(app, token)).Err()
}

func (r *TokenRepository) GetSession(ctx context.Context, app, userID string) (*entity.SessionPointer, error) {
	raw, err := r.rdb.Get(ctx, sessionKey(app, userID)).Result()
	if errors.Is(err, goredis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("token: get session: %w", err)
	}
	var sp entity.SessionPointer
	if err := json.Unmarshal([]byte(raw), &sp); err != nil {
		return nil, fmt.Errorf("token: unmarshal session pointer: %w", err)
	}
	return &sp, nil
}

func (r *TokenRepository) SetSession(ctx context.Context, app, userID string, sp *entity.SessionPointer, ttl time.Duration) error {
	b, err := json.Marshal(sp)
	if err != nil {
		return fmt.Errorf("token: marshal session pointer: %w", err)
	}
	return r.rdb.Set(ctx, sessionKey(app, userID), b, ttl).Err()
}

func (r *TokenRepository) DeleteSession(ctx context.Context, app, userID string) error {
	return r.rdb.Del(ctx, sessionKey(app, userID)).Err()
}

func (r *TokenRepository) GetLoginAttempts(ctx context.Context, app, ip string) (int64, error) {
	n, err := r.rdb.Get(ctx, loginAttemptsKey(app, ip)).Int64()
	if errors.Is(err, goredis.Nil) {
		return 0, nil
	}
	return n, err
}

func (r *TokenRepository) IncrLoginAttempts(ctx context.Context, app, ip string, window time.Duration) (int64, error) {
	key := loginAttemptsKey(app, ip)
	n, err := r.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	// Set TTL only on the first increment so the window is not reset on each failure.
	if n == 1 {
		_ = r.rdb.Expire(ctx, key, window)
	}
	return n, nil
}

func (r *TokenRepository) ResetLoginAttempts(ctx context.Context, app, ip string) error {
	return r.rdb.Del(ctx, loginAttemptsKey(app, ip)).Err()
}

func accessKey(app, token string) string       { return app + ":access:" + token }
func refreshKey(app, token string) string      { return app + ":refresh:" + token }
func sessionKey(app, userID string) string     { return app + ":session:" + userID }
func loginAttemptsKey(app, ip string) string   { return app + ":login_attempts:" + ip }
