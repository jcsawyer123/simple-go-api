package auth

import (
	"time"
)

// NewPermissionCache creates a new cache with specified TTL
func NewPermissionCache(ttl time.Duration) *PermissionCache {
	pc := &PermissionCache{
		cache:   make(map[string]CacheEntry),
		ttl:     ttl,
		cleanup: ttl * 2,
	}

	go pc.startCleanup()
	return pc
}

// startCleanup periodically removes expired entries from the cache
func (pc *PermissionCache) startCleanup() {
	ticker := time.NewTicker(pc.cleanup)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pc.mu.Lock()
			now := time.Now()

			// Check each entry for expiration
			for token, entry := range pc.cache {
				if entry.expiry.Before(now) {
					delete(pc.cache, token)
				}
			}
			pc.mu.Unlock()
		}
	}
}

// Get retrieves permissions from cache if they exist and haven't expired
func (pc *PermissionCache) Get(token string) (map[string]string, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	entry, exists := pc.cache[token]
	if !exists {
		return nil, false
	}

	// Check if the entry has expired
	if entry.expiry.Before(time.Now()) {
		return nil, false
	}

	// Create a copy of the permissions map to prevent external modifications
	permissionsCopy := make(map[string]string, len(entry.permissions))
	for k, v := range entry.permissions {
		permissionsCopy[k] = v
	}

	return permissionsCopy, true
}

// Set stores permissions in cache with expiration time
func (pc *PermissionCache) Set(token string, permissions map[string]string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// Create a copy of the permissions map to prevent external modifications
	permissionsCopy := make(map[string]string, len(permissions))
	for k, v := range permissions {
		permissionsCopy[k] = v
	}

	pc.cache[token] = CacheEntry{
		permissions: permissionsCopy,
		expiry:      time.Now().Add(pc.ttl),
	}
}

// GetParsedPermission retrieves a parsed permission from cache
func (pc *PermissionCache) GetParsedPermission(perm string) (*Permission, bool) {
	if val, ok := pc.parsed.Load(perm); ok {
		return val.(*Permission), true
	}
	return nil, false
}

// SetParsedPermission stores a parsed permission in cache
func (pc *PermissionCache) SetParsedPermission(perm string, p *Permission) {
	pc.parsed.Store(perm, p)
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
