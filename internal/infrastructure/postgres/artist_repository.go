package postgres

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rizkicandra/dandanna-api/internal/domain/entity"
	"github.com/rizkicandra/dandanna-api/internal/domain/repository"
)

type ArtistRepository struct {
	db *sqlx.DB
}

func NewArtistRepository(db *sqlx.DB) *ArtistRepository {
	return &ArtistRepository{db: db}
}

func (r *ArtistRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM "user".user_management WHERE email=$1 AND deleted_at IS NULL)`,
		email,
	).Scan(&exists)
	return exists, err
}

func (r *ArtistRepository) Create(ctx context.Context, params repository.CreateArtistParams) (*entity.Artist, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var (
		userID    string
		createdAt time.Time
		updatedAt time.Time
	)
	err = tx.QueryRowContext(ctx, `
		INSERT INTO "user".user_management (name, phone, email, password, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id, created_at, updated_at`,
		params.Name, params.Phone, params.Email, params.HashedPassword, params.CreatedBy,
	).Scan(&userID, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO "user".artist_profile (user_management_id, business_name, primary_service, city, instagram, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $6)`,
		userID, params.BusinessName, params.PrimaryService, params.City, params.Instagram, params.CreatedBy,
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO "user".user_role (user_management_id, role_id, application_id, status, created_by, updated_by)
		VALUES ($1, $2, $3, 'active', $4, $4)`,
		userID, params.RoleID, params.ApplicationID, params.CreatedBy,
	)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &entity.Artist{
		ID:             userID,
		Name:           params.Name,
		Email:          params.Email,
		Phone:          params.Phone,
		BusinessName:   params.BusinessName,
		PrimaryService: params.PrimaryService,
		City:           params.City,
		Instagram:      params.Instagram,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		CreatedBy:      params.CreatedBy,
	}, nil
}
