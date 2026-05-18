package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
	pginfra "github.com/rizkicandra/dandanna-api/internal/infrastructure/postgres"
	"github.com/rizkicandra/dandanna-api/internal/domain/repository"
)

// testDB opens a connection to the test database.
// Set TEST_DATABASE_URL in the environment before running (e.g. postgres://user:pass@localhost/dandanna_test?sslmode=disable).
// Tests are skipped when the variable is absent.
func testDB(t *testing.T) *sqlx.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping integration tests")
	}
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("failed to ping test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// resolveIDs fetches the application and role IDs seeded by 002_seed_user_schema.sql.
func resolveIDs(t *testing.T, db *sqlx.DB) (appID, roleID int64) {
	t.Helper()
	if err := db.QueryRowContext(context.Background(),
		`SELECT id FROM "user".application WHERE code = 'ARTIST_PORTAL'`,
	).Scan(&appID); err != nil {
		t.Fatalf("failed to resolve ARTIST_PORTAL: %v", err)
	}
	if err := db.QueryRowContext(context.Background(),
		`SELECT id FROM "user".role WHERE code = 'ARTIST'`,
	).Scan(&roleID); err != nil {
		t.Fatalf("failed to resolve ARTIST role: %v", err)
	}
	return
}

// cleanupEmail removes all rows for a test email so tests are idempotent.
func cleanupEmail(t *testing.T, db *sqlx.DB, email string) {
	t.Helper()
	db.MustExecContext(context.Background(), `
		DELETE FROM "user".user_role ur
		USING "user".user_management um
		WHERE ur.user_management_id = um.id AND um.email = $1`, email)
	db.MustExecContext(context.Background(), `
		DELETE FROM "user".artist_profile ap
		USING "user".user_management um
		WHERE ap.user_management_id = um.id AND um.email = $1`, email)
	db.MustExecContext(context.Background(), `
		DELETE FROM "user".user_management WHERE email = $1`, email)
}

func TestArtistRepository_Create_ThreeRows(t *testing.T) {
	db := testDB(t)
	appID, roleID := resolveIDs(t, db)
	repo := pginfra.NewArtistRepository(db)
	email := "integration-test@example.com"
	t.Cleanup(func() { cleanupEmail(t, db, email) })

	params := repository.CreateArtistParams{
		Name:           "Integration User",
		Email:          email,
		Phone:          "081111111111",
		HashedPassword: "$argon2id$v=19$m=65536,t=1,p=4$fakesalt$fakehash",
		BusinessName:   "Test Studio",
		PrimaryService: "makeup",
		City:           "Bandung",
		Instagram:      "@test",
		CreatedBy:      email,
		ApplicationID:  appID,
		RoleID:         roleID,
	}

	artist, err := repo.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if artist.ID == "" {
		t.Error("expected non-empty artist ID")
	}

	var umCount int
	db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM "user".user_management WHERE email = $1`, email,
	).Scan(&umCount)
	if umCount != 1 {
		t.Errorf("user_management rows: got %d, want 1", umCount)
	}

	var apCount int
	db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM "user".artist_profile WHERE user_management_id = $1`, artist.ID,
	).Scan(&apCount)
	if apCount != 1 {
		t.Errorf("artist_profile rows: got %d, want 1", apCount)
	}

	var urCount int
	db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM "user".user_role WHERE user_management_id = $1`, artist.ID,
	).Scan(&urCount)
	if urCount != 1 {
		t.Errorf("user_role rows: got %d, want 1", urCount)
	}
}

func TestArtistRepository_ExistsByEmail_True(t *testing.T) {
	db := testDB(t)
	appID, roleID := resolveIDs(t, db)
	repo := pginfra.NewArtistRepository(db)
	email := "exists-test@example.com"
	t.Cleanup(func() { cleanupEmail(t, db, email) })

	_, err := repo.Create(context.Background(), repository.CreateArtistParams{
		Name:           "Exists Test",
		Email:          email,
		Phone:          "082222222222",
		HashedPassword: "$argon2id$v=19$m=65536,t=1,p=4$fakesalt$fakehash",
		BusinessName:   "Exists Studio",
		PrimaryService: "hair",
		City:           "Surabaya",
		CreatedBy:      email,
		ApplicationID:  appID,
		RoleID:         roleID,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	exists, err := repo.ExistsByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("ExistsByEmail error: %v", err)
	}
	if !exists {
		t.Error("expected ExistsByEmail=true for existing email")
	}
}

func TestArtistRepository_ExistsByEmail_False(t *testing.T) {
	db := testDB(t)
	repo := pginfra.NewArtistRepository(db)

	exists, err := repo.ExistsByEmail(context.Background(), "never-registered@example.com")
	if err != nil {
		t.Fatalf("ExistsByEmail error: %v", err)
	}
	if exists {
		t.Error("expected ExistsByEmail=false for unknown email")
	}
}

func TestArtistRepository_ExistsByEmail_SoftDeleted(t *testing.T) {
	db := testDB(t)
	appID, roleID := resolveIDs(t, db)
	repo := pginfra.NewArtistRepository(db)
	email := "softdelete-test@example.com"
	t.Cleanup(func() { cleanupEmail(t, db, email) })

	artist, err := repo.Create(context.Background(), repository.CreateArtistParams{
		Name:           "Soft Delete",
		Email:          email,
		Phone:          "083333333333",
		HashedPassword: "$argon2id$v=19$m=65536,t=1,p=4$fakesalt$fakehash",
		BusinessName:   "Delete Studio",
		PrimaryService: "attire",
		City:           "Yogyakarta",
		CreatedBy:      email,
		ApplicationID:  appID,
		RoleID:         roleID,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	db.MustExecContext(context.Background(),
		`UPDATE "user".user_management SET deleted_at = NOW() WHERE id = $1`, artist.ID)

	exists, err := repo.ExistsByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("ExistsByEmail error: %v", err)
	}
	if exists {
		t.Error("expected ExistsByEmail=false for soft-deleted email")
	}
}
