package aims

import (
	"sync"

	"github.com/jcsawyer123/simple-go-api/internal/auth/cache"
)

// PermissionCache manages caching of AIMS-specific permissions
type PermissionCache struct {
	cache      *cache.MemoryCache
	parsedPerm sync.Map // Cache for parsed permissions
}

// NewPermissionCache creates a new AIMS permission cache
func NewPermissionCache(cache *cache.MemoryCache) *PermissionCache {
	return &PermissionCache{
		cache: cache,
	}
}

// GetPermissions retrieves permissions from cache if they exist
func (pc *PermissionCache) GetPermissions(token string) (map[string]string, bool) {
	value, found := pc.cache.Get(token)
	if !found {
		return nil, false
	}

	// Type assert to the expected type
	permissions, ok := value.(map[string]string)
	if !ok {
		return nil, false
	}

	// Create a copy to prevent external modifications
	permissionsCopy := make(map[string]string, len(permissions))
	for k, v := range permissions {
		permissionsCopy[k] = v
	}

	return permissionsCopy, true
}

// SetPermissions stores permissions in cache
func (pc *PermissionCache) SetPermissions(token string, permissions map[string]string) {
	// Create a copy to prevent external modifications
	permissionsCopy := make(map[string]string, len(permissions))
	for k, v := range permissions {
		permissionsCopy[k] = v
	}

	pc.cache.Set(token, permissionsCopy)
}

// GetParsedPermission retrieves a parsed permission from cache
func (pc *PermissionCache) GetParsedPermission(perm string) (*Permission, bool) {
	if val, ok := pc.parsedPerm.Load(perm); ok {
		return val.(*Permission), true
	}
	return nil, false
}

// SetParsedPermission stores a parsed permission in cache
func (pc *PermissionCache) SetParsedPermission(perm string, p *Permission) {
	pc.parsedPerm.Store(perm, p)
}

// GetOrParsePerm gets a permission from cache or parses it
func (pc *PermissionCache) GetOrParsePerm(perm string) (*Permission, error) {
	if p, ok := pc.GetParsedPermission(perm); ok {
		return p, nil
	}

	p, err := ParsePermission(perm)
	if err != nil {
		return nil, err
	}

	pc.SetParsedPermission(perm, p)
	return p, nil
}
