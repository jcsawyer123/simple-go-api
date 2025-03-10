package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jcsawyer123/simple-go-api/internal/logger"
	"github.com/sony/gobreaker"
)

// NewClient creates a new auth client
func NewClient(baseURL string) (*AuthClient, error) {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "auth-service",
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

	client := resty.New().
		SetTimeout(5 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(100 * time.Millisecond).
		SetRetryMaxWaitTime(2 * time.Second)

	return &AuthClient{
		baseURL: baseURL,
		client:  client,
		breaker: cb,
		cache:   NewPermissionCache(5 * time.Minute),
	}, nil
}

// ValidateToken validates a token against the auth service
func (c *AuthClient) ValidateToken(ctx context.Context, token string) error {
	_, err := c.breaker.Execute(func() (interface{}, error) {
		resp, err := c.client.R().
			SetContext(ctx).
			SetHeader("x-aims-auth-token", token).
			Get(c.baseURL + "/aims/v1/token_info")

		if err != nil {
			return nil, fmt.Errorf("auth request failed: %w", err)
		}

		if resp.StatusCode() != 200 {
			return nil, fmt.Errorf("invalid token: status %d", resp.StatusCode())
		}

		return resp, nil
	})

	return err
}

// ValidatePermissions checks if the token has the required permission
func (c *AuthClient) ValidatePermissions(ctx context.Context, token, requiredPerm string) error {
	requiredPermObj, err := c.cache.GetOrParsePerm(requiredPerm)
	if err != nil {
		return fmt.Errorf("invalid required permission: %w", err)
	}

	// Check cache first
	if permissions, exists := c.cache.Get(ctx, token); exists {
		logger.InfofWCtx(ctx, "permission check hit cache")
		return checkPermissions(requiredPermObj, permissions)
	}

	logger.WarnfWCtx(ctx, "permission check miss cache")

	// Cache miss - fetch from auth service
	var tokenInfo TokenInfo
	resp, err := c.breaker.Execute(func() (interface{}, error) {
		resp, err := c.client.R().
			SetContext(ctx).
			SetHeader("x-aims-auth-token", token).
			Get(c.baseURL + "/aims/v1/token_info")

		if err != nil {
			return nil, fmt.Errorf("auth request failed: %w", err)
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
	c.cache.Set(ctx, token, allPermissions)

	// Check permissions with the combined set
	return checkPermissions(requiredPermObj, allPermissions)
}
