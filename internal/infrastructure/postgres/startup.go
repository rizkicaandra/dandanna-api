package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// ResolveApplicationID returns the id of the application with the given code.
// Returns an error if the application does not exist or is soft-deleted.
func ResolveApplicationID(ctx context.Context, db *sqlx.DB, code string) (int64, error) {
	var id int64
	err := db.QueryRowContext(ctx,
		`SELECT id FROM "user".application WHERE code = $1 AND deleted_at IS NULL`,
		code,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("resolve application %q: %w", code, err)
	}
	return id, nil
}

// ResolveRoleID returns the id of the role with the given code.
// Returns an error if the role does not exist or is soft-deleted.
func ResolveRoleID(ctx context.Context, db *sqlx.DB, code string) (int64, error) {
	var id int64
	err := db.QueryRowContext(ctx,
		`SELECT id FROM "user".role WHERE code = $1 AND deleted_at IS NULL`,
		code,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("resolve role %q: %w", code, err)
	}
	return id, nil
}
