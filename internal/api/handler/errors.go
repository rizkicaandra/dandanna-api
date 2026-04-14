package handler

import (
	"errors"
	"net/http"

	domainerror "github.com/rizkicandra/dandanna-api/internal/domain/error"
	"github.com/rizkicandra/dandanna-api/internal/api/response"
)

// handleError maps a domain error to the appropriate HTTP status code and
// response envelope. All handlers call this instead of inspecting errors directly.
//
// Mapping:
//   - NotFound       → 404
//   - Conflict       → 409
//   - Validation     → 422
//   - Unprocessable  → 422 (with meta context)
//   - Unauthorized   → 401
//   - Forbidden      → 403
//   - anything else  → 500 (internal detail is never leaked to the client)
func handleError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		notFound      *domainerror.NotFound
		conflict      *domainerror.Conflict
		validation    *domainerror.Validation
		unprocessable *domainerror.Unprocessable
		unauthorized  *domainerror.Unauthorized
		forbidden     *domainerror.Forbidden
	)

	switch {
	case errors.As(err, &notFound):
		response.Error(w, r, http.StatusNotFound,
			response.NewError("", "NOT_FOUND", err.Error()),
		)

	case errors.As(err, &conflict):
		response.Error(w, r, http.StatusConflict,
			response.NewError("", "CONFLICT", err.Error()),
		)

	case errors.As(err, &validation):
		response.Error(w, r, http.StatusUnprocessableEntity,
			response.NewError(validation.Field, validation.Code, validation.Message),
		)

	case errors.As(err, &unprocessable):
		response.Error(w, r, http.StatusUnprocessableEntity,
			response.NewErrorWithMeta(unprocessable.Field, unprocessable.Code, unprocessable.Message, unprocessable.Meta),
		)

	case errors.As(err, &unauthorized):
		response.Error(w, r, http.StatusUnauthorized,
			response.NewError("", "UNAUTHORIZED", err.Error()),
		)

	case errors.As(err, &forbidden):
		response.Error(w, r, http.StatusForbidden,
			response.NewError("", "FORBIDDEN", err.Error()),
		)

	default:
		response.Error(w, r, http.StatusInternalServerError,
			response.NewError("", "INTERNAL_ERROR", "Something went wrong."),
		)
	}
}
