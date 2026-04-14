package dto

// Response is the standard JSON envelope for all API endpoints.
//
// Success example:
//
//	{
//	  "success": true,
//	  "meta": { "requestId": "ee81caa7...", "timestamp": "2026-04-14T08:30:00Z" },
//	  "data": { ... },
//	  "errors": []
//	}
//
// Error example:
//
//	{
//	  "success": false,
//	  "meta": { "requestId": "ee81caa7...", "timestamp": "2026-04-14T08:30:00Z" },
//	  "data": null,
//	  "errors": [{ "field": "bookingDate", "code": "DATE_IN_THE_PAST", "meta": {}, "message": "..." }]
//	}
type Response struct {
	Success bool         `json:"success"`
	Meta    ResponseMeta `json:"meta"`
	Data    interface{}  `json:"data"`
	Errors  []ErrorItem  `json:"errors"`
}

// ResponseMeta contains request-scoped metadata attached to every response.
type ResponseMeta struct {
	RequestID string `json:"requestId"`
	Timestamp string `json:"timestamp"`
}

// ErrorItem describes a single validation or business error.
// Field is optional — omitted for non-field-level errors (e.g. auth failures).
// Meta carries additional context (e.g. required vs current balance).
type ErrorItem struct {
	Field   string                 `json:"field,omitempty"`
	Code    string                 `json:"code"`
	Meta    map[string]interface{} `json:"meta"`
	Message string                 `json:"message"`
}
