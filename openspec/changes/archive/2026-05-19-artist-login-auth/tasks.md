## 1. Infrastructure — Token Generator

- [x] 1.1 Create `internal/infrastructure/token/token.go` with `Generate() string` — 32 random bytes via `crypto/rand`, encoded as 64-char hex using `encoding/hex`

## 2. Domain — Entities & Repository Interface

- [x] 2.1 Create `internal/domain/entity/session.go` — `Session{UserID, Role, Name string}`
- [x] 2.2 Create `internal/domain/entity/session_pointer.go` — `SessionPointer{RefreshToken, AccessToken string}`
- [x] 2.3 Create `internal/domain/repository/token.go` — `TokenRepository` interface with methods: `StoreAccess`, `DeleteAccess`, `StoreRefresh`, `GetRefresh`, `DeleteRefresh`, `GetSession`, `SetSession`, `DeleteSession`, `GetLoginAttempts`, `IncrLoginAttempts`, `ResetLoginAttempts`

## 3. Infrastructure — Redis Token Repository

- [x] 3.1 Create `internal/infrastructure/redis/token_repository.go` implementing `TokenRepository`
- [x] 3.2 Implement `StoreAccess` / `DeleteAccess` using keys `<app>:access:<token>`
- [x] 3.3 Implement `StoreRefresh` / `GetRefresh` / `DeleteRefresh` using keys `<app>:refresh:<token>`
- [x] 3.4 Implement `GetSession` / `SetSession` / `DeleteSession` using keys `<app>:session:<userID>` — value is JSON `SessionPointer`
- [x] 3.5 Implement `GetLoginAttempts` / `IncrLoginAttempts` / `ResetLoginAttempts` using keys `<app>:login_attempts:<ip>` — `IncrLoginAttempts` uses INCR + EXPIRE (set TTL only on first increment)

## 4. Infrastructure — Crypto

- [x] 4.1 Add `VerifyArgon2id(password, hash string) (bool, error)` to `internal/infrastructure/crypto/argon2.go` — parse PHC string, extract salt + params, recompute hash, compare with `crypto/subtle.ConstantTimeCompare`

## 5. Config

- [x] 5.1 Add `AccessTokenTTL time.Duration`, `RefreshTokenTTL time.Duration`, `LoginAttemptWindow time.Duration`, `LoginAttemptLimit int` to `internal/infrastructure/config/config.go`
- [x] 5.2 Add `ACCESS_TOKEN_TTL=15m`, `REFRESH_TOKEN_TTL=168h`, `LOGIN_ATTEMPT_WINDOW=15m`, `LOGIN_ATTEMPT_LIMIT=10` to `.env` and `.env.example`

## 6. Domain — Artist Entity & Repository

- [x] 6.1 Add `HashedPassword string` and `RoleStatus string` fields to `internal/domain/entity/artist.go` (populated only by login query, not registration)
- [x] 6.2 Add `GetByEmail(ctx context.Context, email string) (*entity.Artist, error)` to `internal/domain/repository/artist.go`

## 7. Infrastructure — Postgres Artist Repository

- [x] 7.1 Implement `GetByEmail` in `internal/infrastructure/postgres/artist_repository.go` — JOIN `user_management` + `artist_profile` + `user_role`, scan `password` into `HashedPassword` and `ur.status` into `RoleStatus`, map `sql.ErrNoRows` → `*domainerror.NotFound`

## 8. API — DTOs

- [x] 8.1 Create `internal/api/dto/auth.go` with:
  - `LoginRequest{Email, Password string}` with validate tags `required,email` / `required`
  - `LoginResponse{AccessToken, RefreshToken string, Artist ArtistResponse}`
  - `RefreshRequest{RefreshToken string}` with validate tag `required`
  - `RefreshResponse{AccessToken, RefreshToken string}`

## 9. Application — Artist Service

- [x] 9.1 Add `tokenRepo repository.TokenRepository` dependency to `ArtistService` struct and constructor in `internal/application/service/artist.go`
- [x] 9.2 Implement `Login(ctx, ip string, in LoginInput) (*LoginOutput, error)` — check rate limit → verify credentials → kick old session → issue tokens → update session pointer → reset rate limit counter
- [x] 9.3 Implement `Refresh(ctx, refreshToken string) (*RefreshOutput, error)` — validate refresh token → get session pointer → delete old tokens → issue new pair → update session pointer
- [x] 9.4 Implement `Logout(ctx, refreshToken string) error` — validate refresh token → get session pointer → delete access + refresh tokens → delete session pointer

## 10. API — Handler

- [x] 10.1 Add `Login`, `Refresh`, `Logout` methods to `internal/api/handler/artist.go`
- [x] 10.2 In `Login` handler: extract real IP (`X-Real-IP` header → fallback to `net.SplitHostPort(r.RemoteAddr)`), decode + validate request, call service, return 200 with `LoginResponse`
- [x] 10.3 In `Refresh` handler: decode + validate request, call service, return 200 with `RefreshResponse`
- [x] 10.4 In `Logout` handler: decode + validate request, call service, return 200 with null data
- [x] 10.5 Map `*domainerror.Unauthorized` → 401, `*domainerror.Forbidden` → 403, `*domainerror.Unprocessable` → 422 in `handleError` (verify existing mapping covers these)

## 11. Router

- [x] 11.1 Add `Login(w, r)`, `Refresh(w, r)`, `Logout(w, r)` to the `ArtistHandler` interface in `internal/api/router/router.go`
- [x] 11.2 Register routes: `POST /api/artists/login`, `POST /api/artists/auth/refresh`, `POST /api/artists/logout`

## 12. Bootstrap / Wiring

- [x] 12.1 Instantiate `tokenRepo` (Redis token repository) in `internal/bootstrap/app.go` or `bootstrap/artist.go`
- [x] 12.2 Pass `tokenRepo` to `NewArtistService` in `internal/bootstrap/artist.go`

## 13. Verification

- [x] 13.1 `go build ./...` passes with no errors
- [x] 13.2 `go vet ./...` passes with no warnings
- [x] 13.3 Write table-driven unit tests for `ArtistService.Login` — happy path, wrong password, email not found, suspended account, rate limit exceeded, existing session kicked
- [x] 13.4 Write table-driven unit tests for `ArtistService.Refresh` — happy path, invalid token, already-rotated token
- [x] 13.5 Write table-driven unit tests for `ArtistService.Logout` — happy path, invalid token
- [x] 13.6 `go test -race ./internal/application/service/...` passes
- [x] 13.7 Smoke test: `POST /api/artists/login` → 200 with tokens
- [x] 13.8 Smoke test: `POST /api/artists/auth/refresh` → 200 with new token pair; old refresh token → 401
- [x] 13.9 Smoke test: `POST /api/artists/logout` → 200; access token immediately invalid
- [x] 13.10 Smoke test: 11 failed logins from same IP → 422 `TOO_MANY_LOGIN_ATTEMPTS`
