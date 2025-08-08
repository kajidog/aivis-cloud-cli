package errors

import (
	"fmt"
	"net/http"
)

// APIError represents an error from the Aivis Cloud API
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	RequestID  string `json:"request_id,omitempty"`
}

func (e *APIError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("API error %d: %s (request ID: %s)", e.StatusCode, e.Message, e.RequestID)
	}
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// NewAPIError creates a new API error
func NewAPIError(statusCode int, message string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
	}
}

// NewAPIErrorFromHTTP creates a new API error from HTTP response
func NewAPIErrorFromHTTP(resp *http.Response, message string) *APIError {
	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    message,
		RequestID:  resp.Header.Get("X-Request-ID"),
	}
}

// IsAPIError checks if the error is an API error
func IsAPIError(err error) (*APIError, bool) {
	apiErr, ok := err.(*APIError)
	return apiErr, ok
}

// Common error types
var (
	ErrUnauthorized    = NewAPIError(401, "API key is required or invalid")
	ErrPaymentRequired = NewAPIError(402, "Credit balance is insufficient")
	ErrNotFound        = NewAPIError(404, "Specified model UUID not found")
	ErrUnprocessable   = NewAPIError(422, "Request parameter format is incorrect")
	ErrTooManyRequests = NewAPIError(429, "Rate limit exceeded")
	ErrInternalServer  = NewAPIError(500, "Unknown error occurred during synthesis server connection")
	ErrBadGateway      = NewAPIError(502, "Failed to connect to synthesis server")
	ErrServiceUnavail  = NewAPIError(503, "Synthesis server is experiencing issues")
	ErrGatewayTimeout  = NewAPIError(504, "Connection to synthesis server timed out")
)
