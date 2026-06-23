package manager

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/ab-payment-system/internal/proxy/health"
	"github.com/ab-payment-system/internal/proxy/model"
	"github.com/ab-payment-system/internal/proxy/provider"
)

type ProxyManager struct {
	mu          sync.RWMutex
	proxies     map[string]*model.Proxy   // proxyID -> Proxy
	hotPool     map[string][]*model.Proxy // accountID -> dedicated proxies
	providers   []provider.ProxyProvider
	healthCheck *health.ProxyChecker
	bindingMap  map[string]string // accountID -> proxyID
}

func NewProxyManager(providers []provider.ProxyProvider, hc *health.ProxyChecker) *ProxyManager {
	return &ProxyManager{
		proxies:     make(map[string]*model.Proxy),
		hotPool:     make(map[string][]*model.Proxy),
		providers:   providers,
		healthCheck: hc,
		bindingMap:  make(map[string]string),
	}
}

// AcquireForAccount gets or assigns a proxy for a specific payment account.
func (pm *ProxyManager) AcquireForAccount(ctx context.Context, accountID, country string) (*model.Proxy, error) {
	pm.mu.RLock()

	// Check existing dedicated binding
	if proxyID, ok := pm.bindingMap[accountID]; ok {
		if p, exists := pm.proxies[proxyID]; exists && p.Status == "online" {
			pm.mu.RUnlock()
			return p, nil
		}
	}

	// Check hot pool for this account
	if pool, ok := pm.hotPool[accountID]; ok {
		for _, p := range pool {
			if p.Status == "online" && (country == "" || p.Country == country) {
				pm.mu.RUnlock()
				return p, nil
			}
		}
	}

	pm.mu.RUnlock()

	// Acquire new proxy from providers
	return pm.acquireNew(ctx, accountID, country)
}

func (pm *ProxyManager) acquireNew(ctx context.Context, accountID, country string) (*model.Proxy, error) {
	for _, prov := range pm.providers {
		proxies, err := prov.Fetch(ctx, country, 1)
		if err != nil || len(proxies) == 0 {
			log.Warn().Str("provider", prov.Name()).Err(err).Msg("failed to fetch proxy")
			continue
		}

		p := proxies[0]

		// Health check the new proxy
		proxyHealth := pm.healthCheck.Check(p)
		if !proxyHealth.IsHealthy {
			log.Debug().Str("proxy", p.Host).Str("provider", prov.Name()).Msg("proxy unhealthy")
			continue
		}

		p.BoundAccountID = accountID
		p.Status = "online"
		p.Latency = proxyHealth.Latency

		pm.mu.Lock()
		pm.proxies[p.ID] = p
		pm.bindingMap[accountID] = p.ID
		pm.hotPool[accountID] = append(pm.hotPool[accountID], p)
		pm.mu.Unlock()

		log.Info().Str("proxy", p.ID).Str("account", accountID).Str("provider", prov.Name()).Msg("proxy assigned to account")
		return p, nil
	}

	return nil, fmt.Errorf("no available proxy for account %s", accountID)
}

// Release clears the proxy binding for an account.
func (pm *ProxyManager) Release(accountID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if proxyID, ok := pm.bindingMap[accountID]; ok {
		if p, exists := pm.proxies[proxyID]; exists {
			p.BoundAccountID = ""
		}
	}
	delete(pm.bindingMap, accountID)
	delete(pm.hotPool, accountID)
}

// GetAccountProxy returns the proxy currently bound to an account.
func (pm *ProxyManager) GetAccountProxy(accountID string) (*model.Proxy, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	proxyID, ok := pm.bindingMap[accountID]
	if !ok {
		return nil, false
	}
	p, exists := pm.proxies[proxyID]
	return p, exists
}

// HealthCheckAll runs health checks on all managed proxies.
func (pm *ProxyManager) HealthCheckAll(ctx context.Context) {
	pm.mu.RLock()
	proxies := make([]*model.Proxy, 0, len(pm.proxies))
	for _, p := range pm.proxies {
		proxies = append(proxies, p)
	}
	pm.mu.RUnlock()

	for _, p := range proxies {
		h := pm.healthCheck.Check(p)
		if !h.IsHealthy {
			p.Status = "offline"
			log.Warn().Str("proxy", p.ID).Str("host", p.Host).Msg("proxy marked offline")
		}
		p.Latency = h.Latency
	}
}

// RegisterProxy adds a pre-configured proxy to the pool.
func (pm *ProxyManager) RegisterProxy(p *model.Proxy) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.proxies[p.ID] = p
}

// GetStats returns proxy pool statistics.
func (pm *ProxyManager) GetStats() map[string]int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	stats := map[string]int{
		"total":   len(pm.proxies),
		"online":  0,
		"offline": 0,
		"bound":   len(pm.bindingMap),
	}
	for _, p := range pm.proxies {
		switch p.Status {
		case "online":
			stats["online"]++
		case "offline":
			stats["offline"]++
		}
	}
	return stats
}
