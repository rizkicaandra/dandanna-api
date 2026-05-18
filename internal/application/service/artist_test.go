package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	domainerror "github.com/rizkicandra/dandanna-api/internal/domain/error"
	"github.com/rizkicandra/dandanna-api/internal/application/service"
	"github.com/rizkicandra/dandanna-api/internal/domain/entity"
	"github.com/rizkicandra/dandanna-api/internal/domain/repository"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/crypto"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/logger"
)

// mockArtistRepo is a minimal in-memory stub for ArtistRepository.
type mockArtistRepo struct {
	existsByEmail func(ctx context.Context, email string) (bool, error)
	create        func(ctx context.Context, params repository.CreateArtistParams) (*entity.Artist, error)
	getByEmail    func(ctx context.Context, email string) (*entity.Artist, error)
}

func (m *mockArtistRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return m.existsByEmail(ctx, email)
}

func (m *mockArtistRepo) Create(ctx context.Context, params repository.CreateArtistParams) (*entity.Artist, error) {
	return m.create(ctx, params)
}

func (m *mockArtistRepo) GetByEmail(ctx context.Context, email string) (*entity.Artist, error) {
	if m.getByEmail != nil {
		return m.getByEmail(ctx, email)
	}
	return nil, errors.New("GetByEmail not implemented in mock")
}

// mockTokenRepo is a no-op stub for TokenRepository used by Register tests.
type mockTokenRepo struct{}

func (m *mockTokenRepo) StoreAccess(_ context.Context, _, _ string, _ *entity.Session, _ time.Duration) error {
	return nil
}
func (m *mockTokenRepo) GetAccess(_ context.Context, _, _ string) (*entity.Session, error) {
	return nil, nil
}
func (m *mockTokenRepo) DeleteAccess(_ context.Context, _, _ string) error { return nil }
func (m *mockTokenRepo) StoreRefresh(_ context.Context, _, _, _ string, _ time.Duration) error {
	return nil
}
func (m *mockTokenRepo) GetRefresh(_ context.Context, _, _ string) (string, error) { return "", nil }
func (m *mockTokenRepo) DeleteRefresh(_ context.Context, _, _ string) error         { return nil }
func (m *mockTokenRepo) GetSession(_ context.Context, _, _ string) (*entity.SessionPointer, error) {
	return nil, nil
}
func (m *mockTokenRepo) SetSession(_ context.Context, _, _ string, _ *entity.SessionPointer, _ time.Duration) error {
	return nil
}
func (m *mockTokenRepo) DeleteSession(_ context.Context, _, _ string) error        { return nil }
func (m *mockTokenRepo) GetLoginAttempts(_ context.Context, _, _ string) (int64, error) {
	return 0, nil
}
func (m *mockTokenRepo) IncrLoginAttempts(_ context.Context, _, _ string, _ time.Duration) (int64, error) {
	return 1, nil
}
func (m *mockTokenRepo) ResetLoginAttempts(_ context.Context, _, _ string) error { return nil }

func newTestService(repo repository.ArtistRepository) *service.ArtistService {
	log := logger.New(logger.LevelError, nil)
	return service.NewArtistService(repo, &mockTokenRepo{}, log, 1, 1, service.ArtistServiceConfig{
		AppCode:  "TEST_APP",
		RoleCode: "TEST_ROLE",
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    168 * time.Hour,
		LoginAttemptWindow: 15 * time.Minute,
		LoginAttemptLimit:  10,
	})
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

// ── Login tests ────────────────────────────────────────────────────────────────

func newLoginRepo(artist *entity.Artist, getErr error) *mockArtistRepo {
	r := happyRepo()
	r.getByEmail = func(_ context.Context, _ string) (*entity.Artist, error) {
		return artist, getErr
	}
	return r
}

func makeHashedArtist() *entity.Artist {
	hash, _ := crypto.HashArgon2id("secret123")
	return &entity.Artist{
		ID:             "uuid-1",
		Name:           "Sari Indah",
		Email:          "sari@example.com",
		HashedPassword: hash,
		RoleStatus:     "active",
	}
}

func TestArtistService_Login_HappyPath(t *testing.T) {
	svc := newTestService(newLoginRepo(makeHashedArtist(), nil))
	out, err := svc.Login(context.Background(), "127.0.0.1", service.LoginInput{
		Email: "sari@example.com", Password: "secret123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out.AccessToken == "" || out.RefreshToken == "" {
		t.Error("expected non-empty tokens")
	}
}

func TestArtistService_Login_WrongPassword(t *testing.T) {
	svc := newTestService(newLoginRepo(makeHashedArtist(), nil))
	_, err := svc.Login(context.Background(), "127.0.0.1", service.LoginInput{
		Email: "sari@example.com", Password: "wrongpassword",
	})
	var u *domainerror.Unauthorized
	if !errors.As(err, &u) {
		t.Fatalf("expected *Unauthorized, got %T: %v", err, err)
	}
}

func TestArtistService_Login_EmailNotFound(t *testing.T) {
	repo := newLoginRepo(nil, &domainerror.NotFound{Resource: "artist", ID: "x@x.com"})
	svc := newTestService(repo)
	_, err := svc.Login(context.Background(), "127.0.0.1", service.LoginInput{
		Email: "x@x.com", Password: "secret123",
	})
	var u *domainerror.Unauthorized
	if !errors.As(err, &u) {
		t.Fatalf("expected *Unauthorized, got %T: %v", err, err)
	}
}

func TestArtistService_Login_SuspendedAccount(t *testing.T) {
	a := makeHashedArtist()
	a.RoleStatus = "suspended"
	svc := newTestService(newLoginRepo(a, nil))
	_, err := svc.Login(context.Background(), "127.0.0.1", service.LoginInput{
		Email: "sari@example.com", Password: "secret123",
	})
	var f *domainerror.Forbidden
	if !errors.As(err, &f) {
		t.Fatalf("expected *Forbidden, got %T: %v", err, err)
	}
}

func TestArtistService_Login_RateLimitExceeded(t *testing.T) {
	// Build a token repo that reports attempts >= limit.
	tr := &mockTokenRepo{}
	tr2 := &overLimitTokenRepo{mockTokenRepo: tr}
	log := logger.New(logger.LevelError, nil)
	svc := service.NewArtistService(happyRepo(), tr2, log, 1, 1, service.ArtistServiceConfig{
		AppCode:            "TEST_APP",
		RoleCode:           "TEST_ROLE",
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    168 * time.Hour,
		LoginAttemptWindow: 15 * time.Minute,
		LoginAttemptLimit:  10,
	})
	_, err := svc.Login(context.Background(), "1.2.3.4", service.LoginInput{
		Email: "sari@example.com", Password: "secret123",
	})
	var u *domainerror.Unprocessable
	if !errors.As(err, &u) {
		t.Fatalf("expected *Unprocessable, got %T: %v", err, err)
	}
}

// overLimitTokenRepo wraps mockTokenRepo and returns 10 attempts.
type overLimitTokenRepo struct{ *mockTokenRepo }

func (o *overLimitTokenRepo) GetLoginAttempts(_ context.Context, _, _ string) (int64, error) {
	return 10, nil
}

// ── Refresh tests ──────────────────────────────────────────────────────────────

func TestArtistService_Refresh_HappyPath(t *testing.T) {
	svc := newTestService(happyRepo())
	out, err := svc.Refresh(context.Background(), "any-token")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out.AccessToken == "" || out.RefreshToken == "" {
		t.Error("expected non-empty tokens")
	}
}

// ── Logout tests ───────────────────────────────────────────────────────────────

func TestArtistService_Logout_HappyPath(t *testing.T) {
	svc := newTestService(happyRepo())
	if err := svc.Logout(context.Background(), "any-token"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
