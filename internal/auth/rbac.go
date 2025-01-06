package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

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

// generateCascadingPermissions generates all possible less specific permissions
// For example: "compute:managed:list:prod" would generate:
// ["compute:managed:list:prod", "compute:managed:list:*", "compute:managed:*:*", "compute:*:*:*", "*:*:*:*"]
func (p *Permission) generateCascadingPermissions() []string {
	var perms []string

	// Add the original permission
	perms = append(perms, fmt.Sprintf("%s:%s:%s:%s", p.Service, p.Resource, p.Action, p.Qualifier))

	// Add cascading qualifier
	if p.Qualifier != "*" {
		perms = append(perms, fmt.Sprintf("%s:%s:%s:*", p.Service, p.Resource, p.Action))
	}

	// Add cascading action
	if p.Action != "*" {
		perms = append(perms, fmt.Sprintf("%s:%s:*:*", p.Service, p.Resource))
	}

	// Add cascading resource
	if p.Resource != "*" {
		perms = append(perms, fmt.Sprintf("%s:*:*:*", p.Service))
	}

	// Add fully wildcarded permission
	if p.Service != "*" {
		perms = append(perms, "*:*:*:*")
	}

	return perms
}

// Matches checks if this permission matches the required permission
// Handles wildcards (*) in either permission
func (p *Permission) Matches(required *Permission) bool {
	return (p.Service == "*" || required.Service == "*" || p.Service == required.Service) &&
		(p.Resource == "*" || required.Resource == "*" || p.Resource == required.Resource) &&
		(p.Action == "*" || required.Action == "*" || p.Action == required.Action) &&
		(p.Qualifier == "*" || required.Qualifier == "*" || p.Qualifier == required.Qualifier)
}

// checkPermissions checks if any of the user's permissions match the required permission
// It first checks for explicit denials, then checks for allowed permissions including cascading
func checkPermissions(requiredPerm *Permission, permissions map[string]string) error {
	// First, check for explicit denials
	// Generate all possible forms of the required permission
	possiblePerms := requiredPerm.generateCascadingPermissions()

	// Check if any of these forms are explicitly denied
	for _, permStr := range possiblePerms {
		if status, exists := permissions[permStr]; exists && status == "denied" {
			return fmt.Errorf("permission explicitly denied: %s", permStr)
		}
	}

	// Then check for allowed permissions
	for permStr, status := range permissions {
		if status != "allowed" {
			continue
		}

		permObj, err := ParsePermission(permStr)
		if err != nil {
			continue // Skip invalid permissions
		}

		if permObj.Matches(requiredPerm) {
			return nil // Permission granted
		}
	}

	return fmt.Errorf("permission denied: required %s", requiredPerm)
}

// ValidatePermissions checks if the token has the required permission
func (c *Client) ValidatePermissions(ctx context.Context, token, requiredPerm string) error {
	requiredPermObj, err := ParsePermission(requiredPerm)
	if err != nil {
		return fmt.Errorf("invalid required permission: %w", err)
	}

	// Check cache first
	if permissions, exists := c.cache.Get(token); exists {
		return checkPermissions(requiredPermObj, permissions)
	}

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
	c.cache.Set(token, allPermissions)

	// Check permissions with the combined set
	return checkPermissions(requiredPermObj, allPermissions)
}
