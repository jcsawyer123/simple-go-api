package health

import (
    "context"
    "log"
    "sync"
    "time"
)

type Status string

const (
    StatusHealthy   Status = "healthy"
    StatusDegraded  Status = "degraded"
    StatusUnhealthy Status = "unhealthy"
)

type Check struct {
    Name     string
    Status   Status
    Error    string
    LastCheck time.Time
}

type HealthManager struct {
    mu            sync.RWMutex
    checks        map[string]*Check
    checkInterval time.Duration
    checkers      map[string]HealthChecker
}

type HealthChecker interface {
    Check(ctx context.Context) error
}

func NewHealthManager(checkInterval time.Duration) *HealthManager {
    return &HealthManager{
        checks:        make(map[string]*Check),
        checkers:      make(map[string]HealthChecker),
        checkInterval: checkInterval,
    }
}

func (h *HealthManager) RegisterChecker(name string, checker HealthChecker) {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    h.checkers[name] = checker
    h.checks[name] = &Check{
        Name:   name,
        Status: StatusHealthy,
    }
}

func (h *HealthManager) StartChecks(ctx context.Context) {
    ticker := time.NewTicker(h.checkInterval)
    go func() {
        // Initial check
        h.runChecks(ctx)
        
        for {
            select {
            case <-ctx.Done():
                ticker.Stop()
                return
            case <-ticker.C:
                h.runChecks(ctx)
            }
        }
    }()
}

func (h *HealthManager) runChecks(ctx context.Context) {
    h.mu.Lock()
    defer h.mu.Unlock()

    for name, checker := range h.checkers {
        check := h.checks[name]
        err := checker.Check(ctx)
        
        check.LastCheck = time.Now()
        if err != nil {
            check.Status = StatusUnhealthy
            check.Error = err.Error()
            log.Printf("Health check failed for %s: %v", name, err)
        } else {
            check.Status = StatusHealthy
            check.Error = ""
        }
    }
}

func (h *HealthManager) GetStatus() map[string]Check {
    h.mu.RLock()
    defer h.mu.RUnlock()

    status := make(map[string]Check)
    for name, check := range h.checks {
        status[name] = *check
    }
    return status
}

func (h *HealthManager) IsHealthy() bool {
    h.mu.RLock()
    defer h.mu.RUnlock()

    for _, check := range h.checks {
        if check.Status != StatusHealthy {
            return false
        }
    }
    return true
}
