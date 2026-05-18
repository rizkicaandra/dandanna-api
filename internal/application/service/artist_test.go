package service_test

import (
	"context"
	"errors"
	"testing"

	domainerror "github.com/rizkicandra/dandanna-api/internal/domain/error"
	"github.com/rizkicandra/dandanna-api/internal/domain/entity"
	"github.com/rizkicandra/dandanna-api/internal/domain/repository"
	"github.com/rizkicandra/dandanna-api/internal/application/service"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/logger"
)

// mockArtistRepo is a minimal in-memory stub for ArtistRepository.
type mockArtistRepo struct {
	existsByEmail func(ctx context.Context, email string) (bool, error)
	create        func(ctx context.Context, params repository.CreateArtistParams) (*entity.Artist, error)
}

func (m *mockArtistRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return m.existsByEmail(ctx, email)
}

func (m *mockArtistRepo) Create(ctx context.Context, params repository.CreateArtistParams) (*entity.Artist, error) {
	return m.create(ctx, params)
}

func newTestService(repo repository.ArtistRepository) *service.ArtistService {
	log := logger.New(logger.LevelError, nil)
	return service.NewArtistService(repo, log, 1, 1)
}

var validInput = service.RegisterInput{
	Fullname:       "Sari Indah",
	Email:          "sari@example.com",
	Phone:          "+6281234567890",
	Password:       "secret123",
	BusinessName:   "Sari Beauty",
	PrimaryService: "makeup",
	City:           "Jakarta",
	Instagram:      "@sari",
}

func happyRepo() *mockArtistRepo {
	return &mockArtistRepo{
		existsByEmail: func(_ context.Context, _ string) (bool, error) { return false, nil },
		create: func(_ context.Context, p repository.CreateArtistParams) (*entity.Artist, error) {
			return &entity.Artist{ID: "uuid-1", Name: p.Name, Email: p.Email}, nil
		},
	}
}

func TestArtistService_Register_HappyPath(t *testing.T) {
	svc := newTestService(happyRepo())
	artist, err := svc.Register(context.Background(), validInput)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if artist.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestArtistService_Register_ValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		input     service.RegisterInput
		wantField string
		wantCode  string
	}{
		{
			name:      "missing fullname",
			input:     func() service.RegisterInput { i := validInput; i.Fullname = ""; return i }(),
			wantField: "fullname",
			wantCode:  "REQUIRED",
		},
		{
			name:      "missing email",
			input:     func() service.RegisterInput { i := validInput; i.Email = ""; return i }(),
			wantField: "email",
			wantCode:  "REQUIRED",
		},
		{
			name:      "invalid email",
			input:     func() service.RegisterInput { i := validInput; i.Email = "not-an-email"; return i }(),
			wantField: "email",
			wantCode:  "INVALID_EMAIL",
		},
		{
			name:      "missing phone",
			input:     func() service.RegisterInput { i := validInput; i.Phone = ""; return i }(),
			wantField: "phone",
			wantCode:  "REQUIRED",
		},
		{
			name:      "invalid phone format",
			input:     func() service.RegisterInput { i := validInput; i.Phone = "081234567890"; return i }(),
			wantField: "phone",
			wantCode:  "INVALID_PHONE",
		},
		{
			name:      "missing password",
			input:     func() service.RegisterInput { i := validInput; i.Password = ""; return i }(),
			wantField: "password",
			wantCode:  "REQUIRED",
		},
		{
			name:      "password too short",
			input:     func() service.RegisterInput { i := validInput; i.Password = "short"; return i }(),
			wantField: "password",
			wantCode:  "TOO_SHORT",
		},
		{
			name:      "password too long",
			input:     func() service.RegisterInput { i := validInput; i.Password = string(make([]byte, 101)); return i }(),
			wantField: "password",
			wantCode:  "TOO_LONG",
		},
		{
			name:      "missing business_name",
			input:     func() service.RegisterInput { i := validInput; i.BusinessName = ""; return i }(),
			wantField: "business_name",
			wantCode:  "REQUIRED",
		},
		{
			name:      "missing primary_service",
			input:     func() service.RegisterInput { i := validInput; i.PrimaryService = ""; return i }(),
			wantField: "primary_service",
			wantCode:  "REQUIRED",
		},
		{
			name:      "invalid primary_service",
			input:     func() service.RegisterInput { i := validInput; i.PrimaryService = "nails"; return i }(),
			wantField: "primary_service",
			wantCode:  "INVALID_SERVICE",
		},
		{
			name:      "missing city",
			input:     func() service.RegisterInput { i := validInput; i.City = ""; return i }(),
			wantField: "city",
			wantCode:  "REQUIRED",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newTestService(happyRepo())
			_, err := svc.Register(context.Background(), tc.input)
			var v *domainerror.Validation
			if !errors.As(err, &v) {
				t.Fatalf("expected *domainerror.Validation, got %T: %v", err, err)
			}
			if v.Field != tc.wantField {
				t.Errorf("field: got %q, want %q", v.Field, tc.wantField)
			}
			if v.Code != tc.wantCode {
				t.Errorf("code: got %q, want %q", v.Code, tc.wantCode)
			}
		})
	}
}

func TestArtistService_Register_DuplicateEmail(t *testing.T) {
	repo := happyRepo()
	repo.existsByEmail = func(_ context.Context, _ string) (bool, error) { return true, nil }
	svc := newTestService(repo)

	_, err := svc.Register(context.Background(), validInput)
	var c *domainerror.Conflict
	if !errors.As(err, &c) {
		t.Fatalf("expected *domainerror.Conflict, got %T: %v", err, err)
	}
}

func TestArtistService_Register_RepoError(t *testing.T) {
	repo := happyRepo()
	repo.existsByEmail = func(_ context.Context, _ string) (bool, error) {
		return false, errors.New("db down")
	}
	svc := newTestService(repo)

	_, err := svc.Register(context.Background(), validInput)
	if err == nil {
		t.Fatal("expected an error from repo, got nil")
	}
}
