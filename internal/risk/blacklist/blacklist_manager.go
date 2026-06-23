package blacklist

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type BlacklistManager struct {
	rdb *redis.Client

	mu              sync.RWMutex
	ipBlacklist     map[string]time.Time
	emailBlacklist  map[string]bool
	countryBlacklist map[string]bool
	cardBinBlacklist map[string]bool
}

func NewBlacklistManager(rdb *redis.Client) *BlacklistManager {
	bm := &BlacklistManager{
		rdb:              rdb,
		ipBlacklist:      make(map[string]time.Time),
		emailBlacklist:   make(map[string]bool),
		countryBlacklist: make(map[string]bool),
		cardBinBlacklist: make(map[string]bool),
	}
	go bm.syncFromRedis()
	return bm
}

func (bm *BlacklistManager) IsIPBlocked(ip string) bool {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	if exp, ok := bm.ipBlacklist[ip]; ok {
		if time.Now().Before(exp) {
			return true
		}
	}
	return false
}

func (bm *BlacklistManager) IsEmailBlocked(email string) bool {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.emailBlacklist[strings.ToLower(email)]
}

func (bm *BlacklistManager) IsCountryBlocked(country string) bool {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.countryBlacklist[strings.ToUpper(country)]
}

func (bm *BlacklistManager) IsCardBinBlocked(bin string) bool {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.cardBinBlacklist[bin]
}

func (bm *BlacklistManager) BlockIP(ctx context.Context, ip string, duration time.Duration) error {
	bm.mu.Lock()
	bm.ipBlacklist[ip] = time.Now().Add(duration)
	bm.mu.Unlock()
	return bm.rdb.Set(ctx, "blacklist:ip:"+ip, "true", duration).Err()
}

func (bm *BlacklistManager) BlockEmail(ctx context.Context, email string) error {
	bm.mu.Lock()
	bm.emailBlacklist[strings.ToLower(email)] = true
	bm.mu.Unlock()
	return bm.rdb.Set(ctx, "blacklist:email:"+strings.ToLower(email), "true", 0).Err()
}

func (bm *BlacklistManager) UnblockIP(ctx context.Context, ip string) error {
	bm.mu.Lock()
	delete(bm.ipBlacklist, ip)
	bm.mu.Unlock()
	return bm.rdb.Del(ctx, "blacklist:ip:"+ip).Err()
}

func (bm *BlacklistManager) syncFromRedis() {
	// Periodic sync from Redis to in-memory cache
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		// Sync IP blacklist - rebuild from scratch to pick up deletions
		if keys, err := bm.rdb.Keys(context.Background(), "blacklist:ip:*").Result(); err == nil {
			bm.mu.Lock()
			newMap := make(map[string]time.Time)
			for _, k := range keys {
				ip := strings.TrimPrefix(k, "blacklist:ip:")
				if ttl, _ := bm.rdb.TTL(context.Background(), k).Result(); ttl > 0 {
					newMap[ip] = time.Now().Add(ttl)
				}
			}
			bm.ipBlacklist = newMap
			bm.mu.Unlock()
		}
	}
}
