package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rizkicandra/dandanna-api/internal/domain/entity"
	domainerror "github.com/rizkicandra/dandanna-api/internal/domain/error"
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

func (r *ArtistRepository) GetByEmail(ctx context.Context, email string) (*entity.Artist, error) {
	var row struct {
		ID             string     `db:"id"`
		Name           string     `db:"name"`
		Phone          string     `db:"phone"`
		Email          string     `db:"email"`
		Password       string     `db:"password"`
		CreatedAt      time.Time  `db:"created_at"`
		UpdatedAt      time.Time  `db:"updated_at"`
		DeletedAt      *time.Time `db:"deleted_at"`
		CreatedBy      string     `db:"created_by"`
		BusinessName   string     `db:"business_name"`
		PrimaryService string     `db:"primary_service"`
		City           string     `db:"city"`
		Instagram      string     `db:"instagram"`
		RoleStatus     string     `db:"role_status"`
	}

	err := r.db.QueryRowxContext(ctx, `
		SELECT um.id, um.name, um.phone, um.email, um.password,
		       um.created_at, um.updated_at, um.deleted_at, um.created_by,
		       ap.business_name, ap.primary_service, ap.city, ap.instagram,
		       ur.status AS role_status
		FROM "user".user_management um
		JOIN "user".artist_profile ap ON ap.user_management_id = um.id
		JOIN "user".user_role       ur ON ur.user_management_id = um.id
		WHERE um.email = $1 AND um.deleted_at IS NULL`,
		email,
	).StructScan(&row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, &domainerror.NotFound{Resource: "artist", ID: email}
	}
	if err != nil {
		return nil, err
	}

	return &entity.Artist{
		ID:             row.ID,
		Name:           row.Name,
		Email:          row.Email,
		Phone:          row.Phone,
		BusinessName:   row.BusinessName,
		PrimaryService: row.PrimaryService,
		City:           row.City,
		Instagram:      row.Instagram,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
		DeletedAt:      row.DeletedAt,
		CreatedBy:      row.CreatedBy,
		HashedPassword: row.Password,
		RoleStatus:     row.RoleStatus,
	}, nil
}
