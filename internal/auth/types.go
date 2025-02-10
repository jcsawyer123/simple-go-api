package auth

import (
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sony/gobreaker"
)

// TokenInfo represents the response from token validation
type TokenInfo struct {
	Roles []Role `json:"roles"`
}

// Role represents a user role with associated permissions
type Role struct {
	ID          string                 `json:"id"`
	AccountID   string                 `json:"account_id"`
	Name        string                 `json:"name"`
	Permissions map[string]string      `json:"permissions"`
	Version     int                    `json:"version"`
	Created     map[string]interface{} `json:"created"`
	Modified    map[string]interface{} `json:"modified"`
}

type Permission struct {
	Sections     [MaxSections]string
	UsedSections int
	original     string
}

// AuthClient represents an auth client with caching capability
type AuthClient struct {
	baseURL string
	client  *resty.Client
	breaker *gobreaker.CircuitBreaker
	cache   *PermissionCache
}

// CircuitBreaker interface for circuit breaking functionality
type CircuitBreaker interface {
	Execute(func() (interface{}, error)) (interface{}, error)
}

// CacheEntry represents a cached permission check result
type CacheEntry struct {
	permissions map[string]string
	expiry      time.Time
}

// PermissionCache handles caching of token permissions
type PermissionCache struct {
	mu      sync.RWMutex
	cache   map[string]CacheEntry
	ttl     time.Duration
	cleanup time.Duration
	parsed  sync.Map // Cache for parsed permissions
}
