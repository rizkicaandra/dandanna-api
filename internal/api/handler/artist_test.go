package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	domainerror "github.com/rizkicandra/dandanna-api/internal/domain/error"
	"github.com/rizkicandra/dandanna-api/internal/api/handler"
	"github.com/rizkicandra/dandanna-api/internal/application/service"
	"github.com/rizkicandra/dandanna-api/internal/domain/entity"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/logger"
)

type mockArtistServicer struct {
	register func(ctx context.Context, input service.RegisterInput) (*entity.Artist, error)
	login    func(ctx context.Context, ip string, input service.LoginInput) (*service.LoginOutput, error)
	refresh  func(ctx context.Context, refreshToken string) (*service.RefreshOutput, error)
	logout   func(ctx context.Context, refreshToken string) error
}

func (m *mockArtistServicer) Register(ctx context.Context, input service.RegisterInput) (*entity.Artist, error) {
	return m.register(ctx, input)
}

func (m *mockArtistServicer) Login(ctx context.Context, ip string, input service.LoginInput) (*service.LoginOutput, error) {
	if m.login != nil {
		return m.login(ctx, ip, input)
	}
	return nil, errors.New("Login not implemented in mock")
}

func (m *mockArtistServicer) Refresh(ctx context.Context, refreshToken string) (*service.RefreshOutput, error) {
	if m.refresh != nil {
		return m.refresh(ctx, refreshToken)
	}
	return nil, errors.New("Refresh not implemented in mock")
}

func (m *mockArtistServicer) Logout(ctx context.Context, refreshToken string) error {
	if m.logout != nil {
		return m.logout(ctx, refreshToken)
	}
	return errors.New("Logout not implemented in mock")
}

func newArtistHandler(svc service.ArtistServicer) *handler.Artist {
	log := logger.New(logger.LevelError, nil)
	return handler.NewArtist(log, svc)
}

func postRegister(h *handler.Artist, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/artists/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Register(rr, req)
	return rr
}

var validPayload = map[string]string{
	"fullname":        "Sari Indah",
	"email":           "sari@example.com",
	"phone":           "081234567890",
	"password":        "secret123",
	"business_name":   "Sari Beauty",
	"primary_service": "makeup",
	"city":            "Jakarta",
	"instagram":       "@sari",
}

func marshalJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}
	return b
}

func TestArtistHandler_Register_Created(t *testing.T) {
	svc := &mockArtistServicer{
		register: func(_ context.Context, _ service.RegisterInput) (*entity.Artist, error) {
			return &entity.Artist{
				ID:             "uuid-1",
				Name:           "Sari Indah",
				Email:          "sari@example.com",
				Phone:          "081234567890",
				BusinessName:   "Sari Beauty",
				PrimaryService: "makeup",
				City:           "Jakarta",
				Instagram:      "@sari",
			}, nil
		},
	}

	rr := postRegister(newArtistHandler(svc), marshalJSON(t, validPayload))

	if rr.Code != http.StatusCreated {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusCreated)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["success"] != true {
		t.Errorf("expected success=true")
	}
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data object")
	}
	if data["id"] != "uuid-1" {
		t.Errorf("id: got %v, want uuid-1", data["id"])
	}
}

func TestArtistHandler_Register_Conflict(t *testing.T) {
	svc := &mockArtistServicer{
		register: func(_ context.Context, _ service.RegisterInput) (*entity.Artist, error) {
			return nil, &domainerror.Conflict{Resource: "artist", Message: "email already registered"}
		},
	}

	rr := postRegister(newArtistHandler(svc), marshalJSON(t, validPayload))

	if rr.Code != http.StatusConflict {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusConflict)
	}
}

func TestArtistHandler_Register_ValidationError(t *testing.T) {
	svc := &mockArtistServicer{
		register: func(_ context.Context, _ service.RegisterInput) (*entity.Artist, error) {
			return nil, &domainerror.Validation{Field: "email", Code: "REQUIRED", Message: "email is required"}
		},
	}

	rr := postRegister(newArtistHandler(svc), marshalJSON(t, validPayload))

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	errs, _ := resp["errors"].([]interface{})
	if len(errs) == 0 {
		t.Fatal("expected at least one error item")
	}
	errItem, _ := errs[0].(map[string]interface{})
	if errItem["field"] != "email" {
		t.Errorf("field: got %v, want email", errItem["field"])
	}
	if errItem["code"] != "REQUIRED" {
		t.Errorf("code: got %v, want REQUIRED", errItem["code"])
	}
}

func TestArtistHandler_Register_MalformedBody(t *testing.T) {
	svc := &mockArtistServicer{
		register: func(_ context.Context, _ service.RegisterInput) (*entity.Artist, error) {
			return nil, errors.New("should not be called")
		},
	}

	rr := postRegister(newArtistHandler(svc), []byte(`{not valid json`))

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusBadRequest)
	}
}
