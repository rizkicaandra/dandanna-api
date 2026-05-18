package service

import (
	"context"

	domainerror "github.com/rizkicandra/dandanna-api/internal/domain/error"
	"github.com/rizkicandra/dandanna-api/internal/domain/entity"
	"github.com/rizkicandra/dandanna-api/internal/domain/repository"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/crypto"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/logger"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/validator"
)

type ArtistServicer interface {
	Register(ctx context.Context, input RegisterInput) (*entity.Artist, error)
}

type ArtistService struct {
	repo   repository.ArtistRepository
	log    logger.Logger
	appID  int64
	roleID int64
}

func NewArtistService(repo repository.ArtistRepository, log logger.Logger, appID, roleID int64) *ArtistService {
	return &ArtistService{repo: repo, log: log, appID: appID, roleID: roleID}
}

type RegisterInput struct {
	Fullname       string `json:"fullname"        validate:"required"`
	Email          string `json:"email"           validate:"required,email"`
	Phone          string `json:"phone"           validate:"required,e164"`
	Password       string `json:"password"        validate:"required,min=8,max=100"`
	BusinessName   string `json:"business_name"   validate:"required"`
	PrimaryService string `json:"primary_service" validate:"required,oneof=makeup hair attire"`
	City           string `json:"city"            validate:"required"`
	Instagram      string `json:"instagram"`
}

func (s *ArtistService) Register(ctx context.Context, in RegisterInput) (*entity.Artist, error) {
	if err := validator.Struct(in); err != nil {
		return nil, err
	}

	exists, err := s.repo.ExistsByEmail(ctx, in.Email)
	if err != nil {
		s.log.Error("failed to check email existence", logger.Err(err))
		return nil, err
	}
	if exists {
		return nil, &domainerror.Conflict{Resource: "artist", Message: "email already registered"}
	}

	hashed, err := crypto.HashArgon2id(in.Password)
	if err != nil {
		s.log.Error("failed to hash password", logger.Err(err))
		return nil, err
	}

	artist, err := s.repo.Create(ctx, repository.CreateArtistParams{
		Name:           in.Fullname,
		Email:          in.Email,
		Phone:          in.Phone,
		HashedPassword: hashed,
		BusinessName:   in.BusinessName,
		PrimaryService: in.PrimaryService,
		City:           in.City,
		Instagram:      in.Instagram,
		CreatedBy:      in.Email,
		ApplicationID:  s.appID,
		RoleID:         s.roleID,
	})
	if err != nil {
		s.log.Error("failed to create artist", logger.Err(err))
		return nil, err
	}

	return artist, nil
}
