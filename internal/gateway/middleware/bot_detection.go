package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

type ipWindow struct {
	count     int
	firstSeen time.Time
}

type BotDetector struct {
	mu       sync.RWMutex
	ipCounts map[string]*ipWindow
	banList  map[string]time.Time
}

func NewBotDetector() *BotDetector {
	bd := &BotDetector{
		ipCounts: make(map[string]*ipWindow),
		banList:  make(map[string]time.Time),
	}
	go bd.cleanup()
	return bd
}

func (b *BotDetector) Allow(ip string) bool {
	b.mu.RLock()
	if banExp, banned := b.banList[ip]; banned && time.Now().Before(banExp) {
		b.mu.RUnlock()
		return false
	}
	b.mu.RUnlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	if w, ok := b.ipCounts[ip]; ok {
		if time.Since(w.firstSeen) < time.Minute {
			w.count++
		} else {
			w.count = 1
			w.firstSeen = time.Now()
		}
		if w.count > 60 {
			b.banList[ip] = time.Now().Add(5 * time.Minute)
			delete(b.ipCounts, ip)
			return false
		}
	} else {
		b.ipCounts[ip] = &ipWindow{count: 1, firstSeen: time.Now()}
	}
	return true
}

func (b *BotDetector) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		b.mu.Lock()
		for ip, w := range b.ipCounts {
			if time.Since(w.firstSeen) > 5*time.Minute {
				delete(b.ipCounts, ip)
			}
		}
		for ip, exp := range b.banList {
			if time.Now().After(exp) {
				delete(b.banList, ip)
			}
		}
		b.mu.Unlock()
	}
}

func (b *BotDetector) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
				ip = strings.Split(fwd, ",")[0]
			}
			ip = strings.TrimSpace(ip)
			if !b.Allow(ip) {
				w.Header().Set("Retry-After", "300")
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
