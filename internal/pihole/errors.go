// Package pihole provides an HTTP client for the Pi-hole v6 REST API
// with transparent session-based authentication.
package pihole

import "fmt"

// APIError represents an error response from the Pi-hole API.
type APIError struct {
	StatusCode int    // HTTP status code
	Key        string // Machine-readable error type (e.g. "bad_request")
	Message    string // Human-readable error message
	Hint       string // Additional context (may be empty)
	Endpoint   string // The API path that failed
}

func (e *APIError) Error() string {
	s := fmt.Sprintf("pi-hole API error %d on %s: %s", e.StatusCode, e.Endpoint, e.Message)
	if e.Hint != "" {
		s += " (hint: " + e.Hint + ")"
	}
	return s
}

// AuthError is returned when the Pi-hole API responds with 401 Unauthorized.
type AuthError struct{ *APIError }

// NotFoundError is returned when a requested resource does not exist (404).
type NotFoundError struct{ *APIError }

// ValidationError is returned for invalid request parameters (400).
type ValidationError struct{ *APIError }

// RateLimitError is returned when too many requests are made (429).
type RateLimitError struct{ *APIError }

// classifyError wraps an APIError into the appropriate typed error.
func classifyError(e *APIError) error {
	switch e.StatusCode {
	case 401:
		return &AuthError{e}
	case 404:
		return &NotFoundError{e}
	case 400:
		return &ValidationError{e}
	case 429:
		return &RateLimitError{e}
	default:
		return e
	}
}
