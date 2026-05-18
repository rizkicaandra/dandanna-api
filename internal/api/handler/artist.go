package handler

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/rizkicandra/dandanna-api/internal/api/dto"
	"github.com/rizkicandra/dandanna-api/internal/api/response"
	"github.com/rizkicandra/dandanna-api/internal/application/service"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/logger"
)

const msgMalformedJSON = "malformed JSON body"

// decodeJSON decodes the request body into T and writes a 400 on failure.
// Returns (value, true) on success, (zero, false) if the caller should return immediately.
func decodeJSON[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		response.Error(w, r, http.StatusBadRequest,
			response.NewError("", "BAD_REQUEST", msgMalformedJSON),
		)
		return v, false
	}
	return v, true
}

type ArtistHandler interface {
	Register(http.ResponseWriter, *http.Request)
	Login(http.ResponseWriter, *http.Request)
	Refresh(http.ResponseWriter, *http.Request)
	Logout(http.ResponseWriter, *http.Request)
}

type Artist struct {
	log logger.Logger
	svc service.ArtistServicer
}

func NewArtist(log logger.Logger, svc service.ArtistServicer) *Artist {
	return &Artist{log: log, svc: svc}
}

func (h *Artist) Register(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[dto.RegisterArtistRequest](w, r)
	if !ok {
		return
	}

	artist, err := h.svc.Register(r.Context(), service.RegisterInput{
		Fullname:       req.Fullname,
		Email:          req.Email,
		Phone:          req.Phone,
		Password:       req.Password,
		BusinessName:   req.BusinessName,
		PrimaryService: req.PrimaryService,
		City:           req.City,
		Instagram:      req.Instagram,
	})
	if err != nil {
		handleError(w, r, err)
		return
	}

	if err := response.JSON(w, r, dto.ArtistResponse{
		ID:             artist.ID,
		Fullname:       artist.Name,
		Email:          artist.Email,
		Phone:          artist.Phone,
		BusinessName:   artist.BusinessName,
		PrimaryService: artist.PrimaryService,
		City:           artist.City,
		Instagram:      artist.Instagram,
	}, http.StatusCreated); err != nil {
		h.log.Error("failed to write register artist response", logger.Err(err))
	}
}

func (h *Artist) Login(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[dto.LoginArtistRequest](w, r)
	if !ok {
		return
	}

	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	out, err := h.svc.Login(r.Context(), ip, service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		handleError(w, r, err)
		return
	}

	if err := response.JSON(w, r, dto.LoginArtistResponse{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
		Artist: dto.ArtistResponse{
			ID:             out.Artist.ID,
			Fullname:       out.Artist.Name,
			Email:          out.Artist.Email,
			Phone:          out.Artist.Phone,
			BusinessName:   out.Artist.BusinessName,
			PrimaryService: out.Artist.PrimaryService,
			City:           out.Artist.City,
			Instagram:      out.Artist.Instagram,
		},
	}, http.StatusOK); err != nil {
		h.log.Error("failed to write login response", logger.Err(err))
	}
}

func (h *Artist) Refresh(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[dto.RefreshTokenRequest](w, r)
	if !ok {
		return
	}

	out, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		handleError(w, r, err)
		return
	}

	if err := response.JSON(w, r, dto.RefreshTokenResponse{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	}, http.StatusOK); err != nil {
		h.log.Error("failed to write refresh response", logger.Err(err))
	}
}

func (h *Artist) Logout(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[dto.RefreshTokenRequest](w, r)
	if !ok {
		return
	}

	if err := h.svc.Logout(r.Context(), req.RefreshToken); err != nil {
		handleError(w, r, err)
		return
	}

	if err := response.JSON(w, r, nil, http.StatusOK); err != nil {
		h.log.Error("failed to write logout response", logger.Err(err))
	}
}
