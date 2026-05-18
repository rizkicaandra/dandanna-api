## 1. Dependencies & Setup

- [x] 1.1 Add `golang.org/x/crypto` to `go.mod` via `go get golang.org/x/crypto`
- [x] 1.2 Run `go mod tidy` to update `go.sum`

## 2. Migrations

- [x] 2.1 Create `migrations/001_create_user_schema.sql` — enums (`user.user_role_status`, `user.primary_service`), then all tables in FK dependency order: `application`, `module`, `submodule`, `role`, `permission`, `user_management`, `user_role`, `artist_profile`; all timestamps `timestamptz NOT NULL DEFAULT NOW()`, uuid PKs with `DEFAULT gen_random_uuid()`, `password TEXT`
- [x] 2.2 Add `updated_at` auto-update trigger function and apply it to all tables in `001_create_user_schema.sql`
- [x] 2.3 Create `migrations/002_seed_user_schema.sql` — insert 3 applications (`ARTIST_PORTAL`, `CLIENT_PORTAL`, `BACKOFFICE`) and 4 roles (`ARTIST`, `CLIENT`, `ADMIN`, `SUPERADMIN`) using subqueries on `code` to avoid hardcoded IDs
- [x] 2.4 Run both migrations against local database and verify all tables and seed rows exist

## 3. Domain Layer

- [x] 3.1 Create `internal/domain/entity/artist.go` — `Artist` struct with `ID string` (UUID), `Name`, `Email`, `Phone`, `BusinessName`, `PrimaryService`, `City`, `Instagram`, `CreatedAt`, `UpdatedAt`, `DeletedAt *time.Time`, `CreatedBy`, `UpdatedBy`, `DeletedBy`
- [x] 3.2 Create `internal/domain/repository/artist.go` — `ArtistRepository` interface with `ExistsByEmail(ctx, email) (bool, error)` and `Create(ctx, CreateArtistParams) (*entity.Artist, error)`; define `CreateArtistParams` struct with all fields including `ApplicationID int64`, `RoleID int64`, `CreatedBy string`

## 4. Application Service

- [x] 4.1 Create `internal/application/service/artist.go` — `ArtistServicer` interface and `ArtistService` struct accepting `ArtistRepository`, `logger.Logger`, `appID int64`, `roleID int64`
- [x] 4.2 Implement `RegisterInput` struct and `Register(ctx, RegisterInput) (*entity.Artist, error)` method — validate all required fields, email format, password min length, primary_service enum; return `*domainerror.Validation` on failure
- [x] 4.3 Call `ExistsByEmail` and return `*domainerror.Conflict` if email taken
- [x] 4.4 Hash password with Argon2id (`time=1, memory=65536, threads=4, keyLen=32`), encode as PHC string
- [x] 4.5 Write unit tests for `ArtistService` using a mock `ArtistRepository` — cover happy path, each validation error, and duplicate email

## 5. Infrastructure — Postgres Repository

- [x] 5.1 Create `internal/infrastructure/postgres/artist_repository.go` — `ArtistRepository` struct implementing `repository.ArtistRepository`
- [x] 5.2 Implement `ExistsByEmail` — `SELECT EXISTS(SELECT 1 FROM "user".user_management WHERE email=$1 AND deleted_at IS NULL)`
- [x] 5.3 Implement `Create` — open `sqlx` transaction; INSERT into `user.user_management` with `RETURNING id`; INSERT into `user.artist_profile` using returned UUID; INSERT into `user.user_role` with `status='active'`; commit or rollback on error
- [x] 5.4 Wire `NewArtistRepository(db *sqlx.DB) *ArtistRepository` constructor

## 6. API Layer

- [x] 6.1 Create `internal/api/dto/artist.go` — `RegisterArtistRequest` (8 fields) and `ArtistResponse` (8 fields, `ID string`)
- [x] 6.2 Create `internal/api/handler/artist.go` — `ArtistHandler` interface with `Register(w, r)`; `Artist` struct with `logger` and `ArtistServicer`; decode JSON body, call service, use `handleError()` for domain errors, return `response.JSON(..., http.StatusCreated)` on success
- [x] 6.3 Update `internal/api/router/router.go` — add `ArtistHandler` interface to `Setup()` params, register `POST /api/artists/register`

## 7. Main Wiring

- [x] 7.1 Update `cmd/api/main.go` — at startup query `application.id WHERE code='ARTIST_PORTAL'` and `role.id WHERE code='ARTIST'`; log fatal if either is missing
- [x] 7.2 Instantiate `postgres.NewArtistRepository(db)`, `service.NewArtistService(repo, log, appID, roleID)`, `handler.NewArtist(log, svc)`
- [x] 7.3 Pass artist handler to `r.Setup(health, artist)`

## 8. Tests

- [x] 8.1 Create `internal/application/service/artist_test.go` — table-driven unit tests for `ArtistService` using a mock `ArtistRepository`; cover: happy path, each required field missing, invalid email, password too short, invalid primary_service, duplicate email (conflict)
- [x] 8.2 Create `internal/api/handler/artist_test.go` — unit tests for `Artist.Register` using a mock `ArtistServicer` and `httptest.NewRecorder`; cover: 201 with correct response envelope, 409 conflict, 422 validation error, 400 malformed JSON body
- [x] 8.3 Create `internal/infrastructure/postgres/artist_repository_test.go` — integration tests against a real PostgreSQL instance (run migrations + seed before tests); cover: successful 3-row transaction, `ExistsByEmail` returns true for existing email, `ExistsByEmail` returns false for soft-deleted email, transaction rollback on FK violation
- [x] 8.4 Run `make test` with race detector — all tests pass

## 9. Verification

- [x] 9.1 Run `make build` — binary compiles with no errors
- [x] 9.2 Run `make run` — server starts, logs show resolved `appID` and `roleID`
- [x] 9.3 `POST /api/artists/register` with valid payload → HTTP 201, UUID in `data.id`
- [x] 9.4 Repeat same email → HTTP 409
- [x] 9.5 Missing required field → HTTP 422 with correct `field` and `REQUIRED` code
- [x] 9.6 Invalid `primary_service` → HTTP 422 with `INVALID_SERVICE`
- [x] 9.7 Password < 8 chars → HTTP 422 with `TOO_SHORT`
- [x] 9.8 Verify DB: 3 rows inserted (`user_management`, `artist_profile`, `user_role`), password column is Argon2id PHC string
- [x] 9.9 Run `make test` — all tests pass with race detector
