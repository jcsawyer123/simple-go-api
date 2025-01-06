package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// TokenInfo represents the structure of the token validation response
type TokenInfo struct {
	Roles []Role `json:"roles"`
}

type Role struct {
	ID          string                 `json:"id"`
	AccountID   string                 `json:"account_id"`
	Name        string                 `json:"name"`
	Permissions map[string]string      `json:"permissions"`
	Version     int                    `json:"version"`
	Created     map[string]interface{} `json:"created"`
	Modified    map[string]interface{} `json:"modified"`
}

// Permission represents a structured form of a permission string
type Permission struct {
	Service   string // e.g., "*" or "compute"
	Resource  string // e.g., "managed" or "own"
	Action    string // e.g., "list" or "*"
	Qualifier string // e.g., "*"
}

// ParsePermission converts a permission string into a structured Permission
func ParsePermission(perm string) (*Permission, error) {
	parts := strings.Split(perm, ":")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid permission format: %s", perm)
	}
	return &Permission{
		Service:   parts[0],
		Resource:  parts[1],
		Action:    parts[2],
		Qualifier: parts[3],
	}, nil
}

// Matches checks if this permission matches the required permission
// Handles wildcards (*) in either permission
func (p *Permission) Matches(required *Permission) bool {
	return (p.Service == "*" || required.Service == "*" || p.Service == required.Service) &&
		(p.Resource == "*" || required.Resource == "*" || p.Resource == required.Resource) &&
		(p.Action == "*" || required.Action == "*" || p.Action == required.Action) &&
		(p.Qualifier == "*" || required.Qualifier == "*" || p.Qualifier == required.Qualifier)
}

// ValidatePermissions checks if the token has the required permission
func (c *Client) ValidatePermissions(ctx context.Context, token, requiredPerm string) error {
	// Parse the required permission once, as we'll use it for both cache and API flows
	requiredPermObj, err := ParsePermission(requiredPerm)
	if err != nil {
		return fmt.Errorf("invalid required permission: %w", err)
	}

	// Check cache first
	if permissions, exists := c.cache.Get(token); exists {
		// Check cached permissions
		for permStr, status := range permissions {
			if status != "allowed" {
				continue
			}

			permObj, err := ParsePermission(permStr)
			if err != nil {
				continue // Skip invalid permissions
			}

			if permObj.Matches(requiredPermObj) {
				return nil // Permission granted from cache
			}
		}
		// If we get here, cached permissions didn't match
		return fmt.Errorf("permission denied (cached): required %s", requiredPerm)
	}

	// Cache miss - fetch from auth service
	var tokenInfo TokenInfo

	// Execute the validation request through the circuit breaker
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

	// Parse the response body
	if err := json.Unmarshal(resp.([]byte), &tokenInfo); err != nil {
		return fmt.Errorf("failed to parse token info: %w", err)
	}

	// Combine all permissions from all roles for caching
	allPermissions := make(map[string]string)
	permissionGranted := false

	// Check each role's permissions and build cache
	for _, role := range tokenInfo.Roles {
		for permStr, status := range role.Permissions {
			// Add to combined permissions for cache
			allPermissions[permStr] = status

			// Skip if not allowed or already found a matching permission
			if status != "allowed" || permissionGranted {
				continue
			}

			// Parse the permission from the role
			permObj, err := ParsePermission(permStr)
			if err != nil {
				continue // Skip invalid permissions
			}

			// Check if the permission matches the required one
			if permObj.Matches(requiredPermObj) {
				permissionGranted = true
				// Don't return immediately - continue building cache
			}
		}
	}

	// Cache the permissions regardless of the check result
	c.cache.Set(token, allPermissions)

	if permissionGranted {
		return nil
	}

	return fmt.Errorf("permission denied: required %s", requiredPerm)
}
