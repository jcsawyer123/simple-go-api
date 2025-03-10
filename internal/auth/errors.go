package auth

import (
	"errors"
	"fmt"
)

// Common auth error types
var (
	// ErrInvalidToken indicates the token is not valid
	ErrInvalidToken = errors.New("invalid token")

	// ErrExpiredToken indicates the token has expired
	ErrExpiredToken = errors.New("token expired")

	// ErrInsufficientPermissions indicates the user lacks required permissions
	ErrInsufficientPermissions = errors.New("insufficient permissions")

	// ErrServiceUnavailable indicates the auth service is unavailable
	ErrServiceUnavailable = errors.New("auth service unavailable")

	// ErrPermissionDenied indicates a permission was explicitly denied
	ErrPermissionDenied = errors.New("permission explicitly denied")
)

// AuthError represents an authentication or authorization error with additional context
type AuthError struct {
	// Err is the underlying error
	Err error

	// Message provides additional error context
	Message string

	// StatusCode can be used for HTTP responses
	StatusCode int

	// Details contains additional error details
	Details map[string]interface{}
}

// Error returns the error message
func (e *AuthError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Err.Error()
}

// Unwrap returns the underlying error
func (e *AuthError) Unwrap() error {
	return e.Err
}

// NewAuthError creates a new AuthError
func NewAuthError(err error, message string, statusCode int) *AuthError {
	return &AuthError{
		Err:        err,
		Message:    message,
		StatusCode: statusCode,
		Details:    make(map[string]interface{}),
	}
}

// WithDetail adds a detail to the error
func (e *AuthError) WithDetail(key string, value interface{}) *AuthError {
	e.Details[key] = value
	return e
}

// Is implements the errors.Is interface
func (e *AuthError) Is(target error) bool {
	return errors.Is(e.Err, target)
}
