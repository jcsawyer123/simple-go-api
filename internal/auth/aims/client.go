package aims

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jcsawyer123/simple-go-api/internal/auth"
	"github.com/jcsawyer123/simple-go-api/internal/auth/cache"
	"github.com/jcsawyer123/simple-go-api/internal/logger"
	"github.com/sony/gobreaker"
)

// Register the AIMS implementation with the auth factory
func init() {
	// Register a factory function for creating AIMS middleware
	auth.RegisterMiddlewareFactory(reflect.TypeOf(&Client{}), func(service auth.Service) auth.Middleware {
		// Cast to AIMS client
		aimsClient := service.(*Client)
		return NewMiddleware(aimsClient)
	})
}

// Ensure Client implements the auth.Service interface
var _ auth.Service = (*Client)(nil)

// Client implements the auth.Service interface for AIMS authentication
type Client struct {
	baseURL   string
	client    *resty.Client
	breaker   *gobreaker.CircuitBreaker
	permCache *PermissionCache
}

// NewClient creates a new AIMS auth client with default settings
func NewClient(baseURL string) (*Client, error) {
	// Create circuit breaker
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "aims-auth-service",
		MaxRequests: 3,
		Interval:    0,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Printf("Circuit breaker %s state change: %s -> %s", name, from, to)
		},
	})

	// Create HTTP client
	client := resty.New().
		SetTimeout(5 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(100 * time.Millisecond).
		SetRetryMaxWaitTime(2 * time.Second)

	// Create cache
	memCache := cache.NewMemoryCache(5 * time.Minute)
	permCache := NewPermissionCache(memCache)

	return &Client{
		baseURL:   baseURL,
		client:    client,
		breaker:   cb,
		permCache: permCache,
	}, nil
}

// ValidateToken validates a token against the AIMS auth service
func (c *Client) ValidateToken(ctx context.Context, token string) error {
	_, err := c.breaker.Execute(func() (interface{}, error) {
		resp, err := c.client.R().
			SetContext(ctx).
			SetHeader(AimsHeaderName, token).
			Get(c.baseURL + "/aims/v1/token_info")

		if err != nil {
			return nil, fmt.Errorf("aims auth request failed: %w", err)
		}

		if resp.StatusCode() != 200 {
			return nil, fmt.Errorf("invalid token: status %d", resp.StatusCode())
		}

		return resp, nil
	})

	return err
}

// ValidatePermissions checks if the token has the required permission
func (c *Client) ValidatePermissions(ctx context.Context, token, requiredPerm string) error {
	requiredPermObj, err := c.permCache.GetOrParsePerm(requiredPerm)
	if err != nil {
		return fmt.Errorf("invalid required permission: %w", err)
	}

	// Check cache first
	if permissions, exists := c.permCache.GetPermissions(token); exists {
		logger.InfofWCtx(ctx, "permission check hit cache")
		return CheckPermissions(requiredPermObj, permissions)
	}

	logger.WarnfWCtx(ctx, "permission check miss cache")

	// Cache miss - fetch from auth service
	var tokenInfo TokenInfo
	resp, err := c.breaker.Execute(func() (interface{}, error) {
		resp, err := c.client.R().
			SetContext(ctx).
			SetHeader(AimsHeaderName, token).
			Get(c.baseURL + "/aims/v1/token_info")

		if err != nil {
			return nil, fmt.Errorf("aims auth request failed: %w", err)
		}

		if resp.StatusCode() != 200 {
			return nil, fmt.Errorf("invalid token: status %d", resp.StatusCode())
		}

		return resp.Body(), nil
	})

	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}

	if err := json.Unmarshal(resp.([]byte), &tokenInfo); err != nil {
		return fmt.Errorf("failed to parse token info: %w", err)
	}

	// Combine all permissions from all roles
	allPermissions := make(map[string]string)
	for _, role := range tokenInfo.Roles {
		for permStr, status := range role.Permissions {
			// In case of conflicts, denied takes precedence
			if existing, exists := allPermissions[permStr]; !exists || existing != "denied" {
				allPermissions[permStr] = status
			}
		}
	}

	// Cache the permissions
	c.permCache.SetPermissions(token, allPermissions)

	// Check permissions with the combined set
	return CheckPermissions(requiredPermObj, allPermissions)
}
