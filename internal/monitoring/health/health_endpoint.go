package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
)

type ComponentHealth struct {
	Name      string       `json:"name"`
	Status    HealthStatus `json:"status"`
	Message   string       `json:"message,omitempty"`
	CheckedAt time.Time    `json:"checked_at"`
	LatencyMs int64        `json:"latency_ms"`
}

type SystemHealth struct {
	Status     HealthStatus      `json:"status"`
	Components []ComponentHealth `json:"components"`
	CheckedAt  time.Time         `json:"checked_at"`
}

type HealthChecker interface {
	Name() string
	Check(ctx context.Context) ComponentHealth
}

type HealthRegistry struct {
	mu       sync.RWMutex
	checkers []HealthChecker
}

func NewHealthRegistry() *HealthRegistry {
	return &HealthRegistry{checkers: make([]HealthChecker, 0)}
}

func (hr *HealthRegistry) Register(checker HealthChecker) {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	hr.checkers = append(hr.checkers, checker)
}

func (hr *HealthRegistry) CheckAll(ctx context.Context) SystemHealth {
	hr.mu.RLock()
	checkers := make([]HealthChecker, len(hr.checkers))
	copy(checkers, hr.checkers)
	hr.mu.RUnlock()

	health := SystemHealth{
		Status:    StatusHealthy,
		CheckedAt: time.Now(),
	}

	for _, checker := range checkers {
		start := time.Now()
		component := checker.Check(ctx)
		component.LatencyMs = time.Since(start).Milliseconds()
		component.CheckedAt = time.Now()

		health.Components = append(health.Components, component)

		if component.Status == StatusUnhealthy {
			health.Status = StatusUnhealthy
		} else if component.Status == StatusDegraded && health.Status == StatusHealthy {
			health.Status = StatusDegraded
		}
	}

	return health
}

func (hr *HealthRegistry) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	health := hr.CheckAll(r.Context())
	w.Header().Set("Content-Type", "application/json")
	if health.Status == StatusUnhealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(health)
}

// DBHealthChecker checks database connectivity
type DBHealthChecker struct {
	name string
	ping func(context.Context) error
}

func NewDBHealthChecker(name string, ping func(context.Context) error) *DBHealthChecker {
	return &DBHealthChecker{name: name, ping: ping}
}

func (c *DBHealthChecker) Name() string { return c.name }

func (c *DBHealthChecker) Check(ctx context.Context) ComponentHealth {
	if err := c.ping(ctx); err != nil {
		return ComponentHealth{Name: c.name, Status: StatusUnhealthy, Message: err.Error()}
	}
	return ComponentHealth{Name: c.name, Status: StatusHealthy, Message: "connected"}
}

// RedisHealthChecker checks Redis connectivity
type RedisHealthChecker struct {
	name string
	ping func(context.Context) error
}

func NewRedisHealthChecker(name string, ping func(context.Context) error) *RedisHealthChecker {
	return &RedisHealthChecker{name: name, ping: ping}
}

func (c *RedisHealthChecker) Name() string { return c.name }

func (c *RedisHealthChecker) Check(ctx context.Context) ComponentHealth {
	if err := c.ping(ctx); err != nil {
		return ComponentHealth{Name: c.name, Status: StatusUnhealthy, Message: err.Error()}
	}
	return ComponentHealth{Name: c.name, Status: StatusHealthy, Message: "connected"}
}
