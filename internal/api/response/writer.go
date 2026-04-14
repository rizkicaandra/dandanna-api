package response

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rizkicandra/dandanna-api/internal/api/dto"
	"github.com/rizkicandra/dandanna-api/internal/api/middleware"
)

// JSON writes a successful response envelope.
// The data argument becomes the "data" field; "errors" is always an empty array.
func JSON(w http.ResponseWriter, r *http.Request, data interface{}, status int) error {
	return write(w, dto.Response{
		Success: true,
		Meta:    meta(r),
		Data:    data,
		Errors:  []dto.ErrorItem{},
	}, status)
}

// Error writes a failure response envelope.
// "data" is always null; pass one or more ErrorItem values.
// Build error items with NewError or NewErrorWithMeta.
func Error(w http.ResponseWriter, r *http.Request, status int, errors ...dto.ErrorItem) error {
	return write(w, dto.Response{
		Success: false,
		Meta:    meta(r),
		Data:    nil,
		Errors:  errors,
	}, status)
}

// NewError builds an ErrorItem without extra metadata.
func NewError(field, code, message string) dto.ErrorItem {
	return dto.ErrorItem{
		Field:   field,
		Code:    code,
		Meta:    map[string]interface{}{},
		Message: message,
	}
}

// NewErrorWithMeta builds an ErrorItem with additional context (e.g. required vs current balance).
func NewErrorWithMeta(field, code, message string, meta map[string]interface{}) dto.ErrorItem {
	if meta == nil {
		meta = map[string]interface{}{}
	}
	return dto.ErrorItem{
		Field:   field,
		Code:    code,
		Meta:    meta,
		Message: message,
	}
}

// meta builds the ResponseMeta for the current request.
func meta(r *http.Request) dto.ResponseMeta {
	return dto.ResponseMeta{
		RequestID: middleware.RequestIDFromContext(r.Context()),
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
	}
}

// write encodes and sends the envelope.
func write(w http.ResponseWriter, resp dto.Response, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(resp)
}
