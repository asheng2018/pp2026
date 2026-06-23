package pool

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/ab-payment-system/internal/proxy/health"
	"github.com/ab-payment-system/internal/proxy/model"
)

// Pool manages proxy lifecycle: hot (active), warm (standby), cold (retired).
type Pool struct {
	mu sync.RWMutex

	hot  map[string]*model.Proxy
	warm map[string]*model.Proxy
	cold map[string]*model.Proxy

	checker *health.ProxyChecker

	rotationInterval time.Duration
	maxUseCount      int64
}

func NewPool(checker *health.ProxyChecker, rotationInterval time.Duration, maxUseCount int64) *Pool {
	if rotationInterval <= 0 {
		rotationInterval = 1 * time.Hour
	}
	if maxUseCount <= 0 {
		maxUseCount = 1000
	}
	return &Pool{
		hot:              make(map[string]*model.Proxy),
		warm:             make(map[string]*model.Proxy),
		cold:             make(map[string]*model.Proxy),
		checker:          checker,
		rotationInterval: rotationInterval,
		maxUseCount:      maxUseCount,
	}
}

// AddToWarm adds a proxy to the warm pool for pre-warming.
func (p *Pool) AddToWarm(px *model.Proxy) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.warm[px.ID] = px
	log.Debug().Str("proxy", px.ID).Msg("proxy added to warm pool")
}

// PromoteToHot moves a proxy from warm to hot pool.
func (p *Pool) PromoteToHot(proxyID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	px, ok := p.warm[proxyID]
	if !ok {
		return nil
	}
	delete(p.warm, proxyID)
	px.Status = "online"
	p.hot[proxyID] = px
	log.Info().Str("proxy", proxyID).Msg("proxy promoted to hot pool")
	return nil
}

// AcquireHot returns an available proxy from the hot pool.
func (p *Pool) AcquireHot(country string) (*model.Proxy, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, px := range p.hot {
		if px.Status == "online" && (country == "" || px.Country == country) {
			return px, true
		}
	}
	return nil, false
}

// Retire moves a proxy to the cold pool.
func (p *Pool) Retire(proxyID, reason string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if px, ok := p.hot[proxyID]; ok {
		delete(p.hot, proxyID)
		px.Status = "banned"
		p.cold[proxyID] = px
		log.Warn().Str("proxy", proxyID).Str("reason", reason).Msg("proxy retired to cold pool")
		return
	}
	if px, ok := p.warm[proxyID]; ok {
		delete(p.warm, proxyID)
		px.Status = "banned"
		p.cold[proxyID] = px
	}
}

// HealthCheck runs health checks on hot and warm pools.
func (p *Pool) HealthCheck(ctx context.Context) {
	p.mu.RLock()
	all := make([]*model.Proxy, 0, len(p.hot)+len(p.warm))
	for _, px := range p.hot {
		all = append(all, px)
	}
	for _, px := range p.warm {
		all = append(all, px)
	}
	p.mu.RUnlock()

	for _, px := range all {
		h := p.checker.Check(px)
		if !h.IsHealthy {
			log.Warn().Str("proxy", px.ID).Msg("proxy failed health check, retiring")
			p.Retire(px.ID, "health_check_failed")
			continue
		}
		px.Latency = h.Latency

		// Rotate if overused
		if px.UsedCount > p.maxUseCount {
			p.Retire(px.ID, "max_use_count_exceeded")
		}
	}
}

// Stats returns pool statistics.
func (p *Pool) Stats() map[string]int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return map[string]int{
		"hot":  len(p.hot),
		"warm": len(p.warm),
		"cold": len(p.cold),
	}
}
