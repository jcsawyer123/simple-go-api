package aims

import (
	"fmt"
	"sort"
	"strings"
)

const (
	// MaxSections is the maximum number of sections in a permission string
	MaxSections = 5

	// WildcardValue represents the wildcard permission symbol
	WildcardValue = "*"

	// AimsHeaderName is the header name for AIMS tokens
	AimsHeaderName = "x-aims-auth-token"

	// Common AIMS permission constants
	MyServiceUpdatePerm          = "myservice:managed:update:*"
	InstigatorDisableAccountPerm = "instigator:*:disable:account"
)

// Permission represents a structured AIMS permission with sections
type Permission struct {
	Sections     [MaxSections]string
	UsedSections int
	original     string
}

// ParsePermission converts a permission string into a structured Permission
func ParsePermission(perm string) (*Permission, error) {
	if perm == WildcardValue {
		return &Permission{
			Sections:     [MaxSections]string{WildcardValue},
			UsedSections: 1,
			original:     WildcardValue,
		}, nil
	}

	parts := strings.Split(perm, ":")
	if len(parts) > MaxSections {
		return nil, fmt.Errorf("invalid permission format (too many parts): %s", perm)
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid permission format (empty): %s", perm)
	}

	p := &Permission{
		UsedSections: len(parts),
		original:     perm,
	}

	// Fill the sections array
	for i := 0; i < MaxSections; i++ {
		if i < len(parts) {
			// If part is empty string, keep it as empty
			p.Sections[i] = parts[i]
		} else {
			// For unused sections, use wildcard
			p.Sections[i] = WildcardValue
		}
	}

	return p, nil
}

// String returns the string representation of the permission
func (p *Permission) String() string {
	if p.original != "" {
		return p.original
	}

	// Build string only up to used sections
	parts := make([]string, p.UsedSections)

	for i := 0; i < p.UsedSections; i++ {
		parts[i] = p.Sections[i]
	}
	return strings.Join(parts, ":")
}

// isMoreSpecificThan checks if this permission is more specific than the other permission
func (p *Permission) isMoreSpecificThan(other *Permission) bool {
	// If other is "*", this is always more specific
	if other.UsedSections == 1 && other.Sections[0] == WildcardValue {
		return p.UsedSections > 1 || p.Sections[0] != WildcardValue
	}

	// If used sections are different, more sections is more specific
	if p.UsedSections != other.UsedSections {
		return p.UsedSections > other.UsedSections
	}

	// Count non-wildcards in used sections
	pCount, oCount := 0, 0
	for i := 0; i < p.UsedSections; i++ {
		if p.Sections[i] != WildcardValue && p.Sections[i] != "" {
			pCount++
		}
		if other.Sections[i] != WildcardValue && other.Sections[i] != "" {
			oCount++
		}
	}
	return pCount > oCount
}

// Matches checks if this permission matches the required permission
func (p *Permission) Matches(required *Permission) bool {
	// Fast path for exact matches
	if p.original == required.original {
		return true
	}

	// Fast path for single "*"
	if (p.UsedSections == 1 && p.Sections[0] == WildcardValue) ||
		(required.UsedSections == 1 && required.Sections[0] == WildcardValue) {
		return true
	}

	// Compare sections
	maxSections := p.UsedSections
	if required.UsedSections > maxSections {
		maxSections = required.UsedSections
	}

	for i := 0; i < maxSections; i++ {
		pSection := p.Sections[i]
		rSection := required.Sections[i]

		// Handle empty sections as wildcards
		if pSection == "" {
			pSection = WildcardValue
		}
		if rSection == "" {
			rSection = WildcardValue
		}

		// If neither is wildcard and they don't match, fail
		if pSection != WildcardValue && rSection != WildcardValue && pSection != rSection {
			return false
		}
	}

	return true
}

// CheckPermissions checks if any of the user's permissions match the required permission
func CheckPermissions(requiredPerm *Permission, permissions map[string]string) error {
	// First check explicit denials
	for permStr, status := range permissions {
		if status != "denied" {
			continue
		}

		deniedPerm, err := ParsePermission(permStr)
		if err != nil {
			continue
		}

		if deniedPerm.Matches(requiredPerm) && !requiredPerm.isMoreSpecificThan(deniedPerm) {
			return fmt.Errorf("permission explicitly denied: %s", permStr)
		}
	}

	// Then check for allowed permissions
	var allowedPerms []*Permission
	for permStr, status := range permissions {
		if status != "allowed" {
			continue
		}

		permObj, err := ParsePermission(permStr)
		if err != nil {
			continue
		}
		allowedPerms = append(allowedPerms, permObj)
	}

	// Sort by specificity
	sort.Slice(allowedPerms, func(i, j int) bool {
		return allowedPerms[i].isMoreSpecificThan(allowedPerms[j])
	})

	// Check permissions from most specific to least specific
	for _, permObj := range allowedPerms {
		if permObj.Matches(requiredPerm) {
			return nil // Permission granted
		}
	}

	return fmt.Errorf("permission denied: required %s", requiredPerm)
}
