package auth

import (
	"time"
)

// BaseServiceConfig contains common configuration for auth services
type BaseServiceConfig struct {
	// Timeout is the request timeout duration
	Timeout time.Duration
	// RetryCount is the number of retry attempts
	RetryCount int
	// CacheTTL is the time-to-live for cached permissions
	CacheTTL time.Duration
	// CircuitBreakerMaxRequests is the maximum number of requests allowed when the circuit breaker is half-open
	CircuitBreakerMaxRequests uint32
	// CircuitBreakerTimeout is the timeout after which the circuit breaker will transition from open to half-open
	CircuitBreakerTimeout time.Duration

	CircuitBreakerFailureThreshold float64
}

// DefaultServiceConfig returns a BaseServiceConfig with sensible defaults
func DefaultServiceConfig() BaseServiceConfig {
	return BaseServiceConfig{
		Timeout:                        5 * time.Second,
		RetryCount:                     3,
		CacheTTL:                       5 * time.Minute,
		CircuitBreakerMaxRequests:      3,
		CircuitBreakerTimeout:          10 * time.Second,
		CircuitBreakerFailureThreshold: 0.6,
	}
}

// AuthResult represents the result of an authentication check
type AuthResult struct {
	// Authenticated indicates if the authentication was successful
	Authenticated bool
	// Error provides information about authentication failures
	Error error
	// Permissions contains the user's permissions (format depends on implementation)
	Permissions interface{}
	// Metadata contains additional auth-related information
	Metadata map[string]interface{}
}
