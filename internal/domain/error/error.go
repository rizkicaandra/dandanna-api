package domainerror

import "fmt"

// NotFound is returned when a requested resource does not exist.
type NotFound struct {
	Resource string
	ID       string
}

func (e *NotFound) Error() string {
	return fmt.Sprintf("%s with id %s not found", e.Resource, e.ID)
}

// Conflict is returned when an operation violates a uniqueness or state constraint.
type Conflict struct {
	Resource string
	Message  string
}

func (e *Conflict) Error() string {
	return fmt.Sprintf("conflict on %s: %s", e.Resource, e.Message)
}

// Validation is returned when input violates a business rule.
// Field and Code are used directly in the response envelope.
type Validation struct {
	Field   string
	Code    string
	Message string
}

func (e *Validation) Error() string {
	return fmt.Sprintf("validation failed on %s: %s", e.Field, e.Message)
}

// Unauthorized is returned when the caller is not authenticated.
type Unauthorized struct {
	Message string
}

func (e *Unauthorized) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "unauthorized"
}

// Forbidden is returned when the caller is authenticated but lacks permission.
type Forbidden struct {
	Message string
}

func (e *Forbidden) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "forbidden"
}

// Unprocessable is returned when a request is valid but cannot be executed
// due to business state (e.g. insufficient balance, already booked slot).
// Meta carries additional context surfaced directly to the client.
type Unprocessable struct {
	Field   string
	Code    string
	Message string
	Meta    map[string]interface{}
}

func (e *Unprocessable) Error() string {
	return fmt.Sprintf("unprocessable: %s", e.Message)
}
