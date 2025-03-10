package cache

import (
	"context"
	"time"
)

// Service defines the generic cache service interface
type Service interface {
	// Get retrieves a value from the cache
	Get(ctx context.Context, key string) (interface{}, bool)

	// Set stores a value in the cache with a TTL
	Set(ctx context.Context, key string, value interface{})

	// Delete removes a value from the cache
	Delete(ctx context.Context, key string)

	// Clear removes all values from the cache
	Clear(ctx context.Context)
}

// Config contains configuration for cache services
type Config struct {
	// TTL is the default time-to-live for cache entries
	TTL time.Duration

	// CleanupInterval is how often the cache is checked for expired entries
	CleanupInterval time.Duration
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		TTL:             5 * time.Minute,
		CleanupInterval: 10 * time.Minute,
	}
}
