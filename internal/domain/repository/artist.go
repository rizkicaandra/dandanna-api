package repository

import (
	"context"

	"github.com/rizkicandra/dandanna-api/internal/domain/entity"
)

type CreateArtistParams struct {
	Name           string
	Email          string
	Phone          string
	HashedPassword string
	BusinessName   string
	PrimaryService string
	City           string
	Instagram      string
	CreatedBy      string
	ApplicationID  int64
	RoleID         int64
}

type ArtistRepository interface {
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Create(ctx context.Context, params CreateArtistParams) (*entity.Artist, error)
}
