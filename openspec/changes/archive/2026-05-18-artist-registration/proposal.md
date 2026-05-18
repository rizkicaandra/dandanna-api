## Why

The platform needs an onboarding flow for professional artists (makeup, hair, attire) — the supply side of the marketplace. Without artist accounts, no bookings can exist. This is the foundational feature that enables all subsequent artist-facing functionality.

## What Changes

- New `POST /api/artists/register` endpoint accepting professional profile data
- Full `user` schema migration — all tables are currently empty; this introduces the schema for the first time
- Seed data for applications (`ARTIST_PORTAL`, `CLIENT_PORTAL`, `BACKOFFICE`) and roles (`ARTIST`, `CLIENT`, `ADMIN`, `SUPERADMIN`)
- Password hashed with Argon2id before storage; no plaintext ever persisted
- Registration creates three rows atomically: `user.user_management` + `user.artist_profile` + `user.user_role`

## Capabilities

### New Capabilities

- `artist-registration`: Self-service registration for professional artists with identity and profile data, resulting in an active artist account

### Modified Capabilities

_(none — first feature, no existing specs)_

## Impact

- **New files**: migrations (×2), entity, repository interface, application service, postgres repository, DTO, handler
- **Modified files**: `internal/api/router/router.go`, `cmd/api/main.go`
- **New dependency**: `golang.org/x/crypto` (Argon2id)
- **Database**: Full `user` schema created for the first time; seed data required before endpoint is usable
- **No breaking changes** — first endpoint beyond health/readyz
