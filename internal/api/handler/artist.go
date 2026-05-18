package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rizkicandra/dandanna-api/internal/api/dto"
	"github.com/rizkicandra/dandanna-api/internal/api/response"
	"github.com/rizkicandra/dandanna-api/internal/application/service"
	"github.com/rizkicandra/dandanna-api/internal/infrastructure/logger"
)

type ArtistHandler interface {
	Register(http.ResponseWriter, *http.Request)
}

type Artist struct {
	log logger.Logger
	svc service.ArtistServicer
}

func NewArtist(log logger.Logger, svc service.ArtistServicer) *Artist {
	return &Artist{log: log, svc: svc}
}

func (h *Artist) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterArtistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, http.StatusBadRequest,
			response.NewError("", "BAD_REQUEST", "malformed JSON body"),
		)
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
