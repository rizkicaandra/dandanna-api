package service

import (
	"context"
	"strings"
	"time"

	domainerror "github.com/rizkicandra/dandanna-api/internal/domain/error"
	"github.com/rizkicandra/dandanna-api/internal/domain/entity"
	"github.com/rizkicandra/dandanna-api/internal/domain/repository"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/crypto"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/logger"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/token"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/validator"
)

type ArtistServicer interface {
	Register(ctx context.Context, input RegisterInput) (*entity.Artist, error)
	Login(ctx context.Context, ip string, in LoginInput) (*LoginOutput, error)
	Refresh(ctx context.Context, refreshToken string) (*RefreshOutput, error)
	Logout(ctx context.Context, refreshToken string) error
}

type ArtistService struct {
	repo      repository.ArtistRepository
	tokenRepo repository.TokenRepository
	log       logger.Logger
	appCode   string
	roleCode  string
	appID     int64
	roleID    int64
	cfg       ArtistServiceConfig
}

// ArtistServiceConfig carries all startup configuration for ArtistService.
type ArtistServiceConfig struct {
	AppCode            string
	RoleCode           string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	LoginAttemptWindow time.Duration
	LoginAttemptLimit  int
}

func NewArtistService(
	repo repository.ArtistRepository,
	tokenRepo repository.TokenRepository,
	log logger.Logger,
	appID, roleID int64,
	cfg ArtistServiceConfig,
) *ArtistService {
	return &ArtistService{
		repo:      repo,
		tokenRepo: tokenRepo,
		log:       log,
		appCode:   strings.ToLower(cfg.AppCode),
		roleCode:  cfg.RoleCode,
		appID:     appID,
		roleID:    roleID,
		cfg:       cfg,
	}
}

// ── Login / Refresh / Logout types ────────────────────────────────────────────

type LoginInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type LoginOutput struct {
	Artist       *entity.Artist
	AccessToken  string
	RefreshToken string
}

type RefreshOutput struct {
	AccessToken  string
	RefreshToken string
}

// ── Register types ─────────────────────────────────────────────────────────────

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

func (s *ArtistService) Login(ctx context.Context, ip string, in LoginInput) (*LoginOutput, error) {
	if err := validator.Struct(in); err != nil {
		return nil, err
	}

	attempts, err := s.tokenRepo.GetLoginAttempts(ctx, s.appCode, ip)
	if err != nil {
		s.log.Error("failed to get login attempts", logger.Err(err))
		return nil, err
	}
	if attempts >= int64(s.cfg.LoginAttemptLimit) {
		return nil, &domainerror.Unprocessable{
			Code:    "TOO_MANY_LOGIN_ATTEMPTS",
			Message: "too many login attempts, try again later",
		}
	}

	artist, err := s.repo.GetByEmail(ctx, in.Email)
	if err != nil {
		_, _ = s.tokenRepo.IncrLoginAttempts(ctx, s.appCode, ip, s.cfg.LoginAttemptWindow)
		return nil, &domainerror.Unauthorized{Message: "invalid email or password"}
	}

	ok, err := crypto.VerifyArgon2id(in.Password, artist.HashedPassword)
	if err != nil || !ok {
		_, _ = s.tokenRepo.IncrLoginAttempts(ctx, s.appCode, ip, s.cfg.LoginAttemptWindow)
		return nil, &domainerror.Unauthorized{Message: "invalid email or password"}
	}

	if artist.RoleStatus != "active" {
		return nil, &domainerror.Forbidden{Message: "account is suspended"}
	}

	// Kick existing session before issuing new tokens.
	existing, err := s.tokenRepo.GetSession(ctx, s.appCode, artist.ID)
	if err != nil {
		s.log.Error("failed to get existing session", logger.Err(err))
		return nil, err
	}
	if existing != nil {
		_ = s.tokenRepo.DeleteRefresh(ctx, s.appCode, existing.RefreshToken)
		_ = s.tokenRepo.DeleteAccess(ctx, s.appCode, existing.AccessToken)
	}

	accessToken := token.Generate()
	refreshToken := token.Generate()

	if err := s.tokenRepo.StoreAccess(ctx, s.appCode, accessToken, &entity.Session{
		UserID: artist.ID,
		Role:   s.roleCode,
		Name:   artist.Name,
	}, s.cfg.AccessTokenTTL); err != nil {
		s.log.Error("failed to store access token", logger.Err(err))
		return nil, err
	}
	if err := s.tokenRepo.StoreRefresh(ctx, s.appCode, refreshToken, artist.ID, s.cfg.RefreshTokenTTL); err != nil {
		s.log.Error("failed to store refresh token", logger.Err(err))
		return nil, err
	}
	if err := s.tokenRepo.SetSession(ctx, s.appCode, artist.ID, &entity.SessionPointer{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}, s.cfg.RefreshTokenTTL); err != nil {
		s.log.Error("failed to set session pointer", logger.Err(err))
		return nil, err
	}

	_ = s.tokenRepo.ResetLoginAttempts(ctx, s.appCode, ip)

	return &LoginOutput{Artist: artist, AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *ArtistService) Refresh(ctx context.Context, refreshToken string) (*RefreshOutput, error) {
	userID, err := s.tokenRepo.GetRefresh(ctx, s.appCode, refreshToken)
	if err != nil {
		return nil, err
	}

	sp, err := s.tokenRepo.GetSession(ctx, s.appCode, userID)
	if err != nil {
		s.log.Error("failed to get session pointer on refresh", logger.Err(err))
		return nil, err
	}

	// Read session data from the current access token before deleting it
	// so we can populate the new access token without a DB round-trip.
	var session *entity.Session
	if sp != nil {
		session, _ = s.tokenRepo.GetAccess(ctx, s.appCode, sp.AccessToken)
		_ = s.tokenRepo.DeleteAccess(ctx, s.appCode, sp.AccessToken)
	}
	if session == nil {
		session = &entity.Session{UserID: userID, Role: s.roleCode}
	}

	_ = s.tokenRepo.DeleteRefresh(ctx, s.appCode, refreshToken)

	newAccess := token.Generate()
	newRefresh := token.Generate()

	if err := s.tokenRepo.StoreAccess(ctx, s.appCode, newAccess, session, s.cfg.AccessTokenTTL); err != nil {
		s.log.Error("failed to store new access token", logger.Err(err))
		return nil, err
	}
	if err := s.tokenRepo.StoreRefresh(ctx, s.appCode, newRefresh, userID, s.cfg.RefreshTokenTTL); err != nil {
		s.log.Error("failed to store new refresh token", logger.Err(err))
		return nil, err
	}
	if err := s.tokenRepo.SetSession(ctx, s.appCode, userID, &entity.SessionPointer{
		RefreshToken: newRefresh,
		AccessToken:  newAccess,
	}, s.cfg.RefreshTokenTTL); err != nil {
		s.log.Error("failed to update session pointer", logger.Err(err))
		return nil, err
	}

	return &RefreshOutput{AccessToken: newAccess, RefreshToken: newRefresh}, nil
}

func (s *ArtistService) Logout(ctx context.Context, refreshToken string) error {
	userID, err := s.tokenRepo.GetRefresh(ctx, s.appCode, refreshToken)
	if err != nil {
		return err
	}

	existing, err := s.tokenRepo.GetSession(ctx, s.appCode, userID)
	if err != nil {
		s.log.Error("failed to get session pointer on logout", logger.Err(err))
		return err
	}

	_ = s.tokenRepo.DeleteRefresh(ctx, s.appCode, refreshToken)
	if existing != nil {
		_ = s.tokenRepo.DeleteAccess(ctx, s.appCode, existing.AccessToken)
	}
	_ = s.tokenRepo.DeleteSession(ctx, s.appCode, userID)

	return nil
}
