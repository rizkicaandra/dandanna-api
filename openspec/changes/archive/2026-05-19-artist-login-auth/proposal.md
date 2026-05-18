## Why

Artists can register but have no way to log in — the platform is unusable without authentication. This change introduces a secure, session-aware login flow so artists can access protected endpoints.

## What Changes

- New `POST /api/artists/login` endpoint: validates credentials, enforces IP-based brute-force protection, issues a pair of opaque Redis-backed tokens, and kicks any existing session (single-session policy)
- New `POST /api/artists/auth/refresh` endpoint: rotates both access and refresh tokens on every call (single-use refresh tokens)
- New `POST /api/artists/logout` endpoint: immediately revokes both tokens from Redis — no waiting for TTL expiry
- New token infrastructure: opaque 64-char hex tokens (32 random bytes), stored under app-namespaced Redis keys (`artist:*`) — cross-app token misuse is structurally impossible
- New IP-based rate limiting on login: max 10 failed attempts per 15-minute window per IP

## Capabilities

### New Capabilities
- `artist-auth`: Login, refresh, and logout for artists — token issuance, rotation, and revocation backed by Redis with single-session enforcement and brute-force protection

### Modified Capabilities

## Impact

- **New routes**: `POST /api/artists/login`, `POST /api/artists/auth/refresh`, `POST /api/artists/logout`
- **New Redis keys**: `artist:access:<token>`, `artist:refresh:<token>`, `artist:session:<userID>`, `artist:login_attempts:<ip>`
- **New files**: token generator, Session entity, SessionPointer entity, TokenRepository interface, Redis token repository, auth DTOs
- **Modified files**: `crypto/argon2.go`, `config/config.go`, `domain/entity/artist.go`, `domain/repository/artist.go`, `postgres/artist_repository.go`, `application/service/artist.go`, `api/handler/artist.go`, `api/router/router.go`, `bootstrap/artist.go`, `bootstrap/app.go`, `.env`
- **Dependencies unchanged**: no new Go packages required — uses existing `crypto/rand`, `encoding/hex`, `crypto/subtle`, `go-redis`
