package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"
	"sync"
)

type OriginChecker struct {
	mu             sync.RWMutex
	allowedOrigins map[string]bool
	allowedASites  map[string]string // domain -> merchant_id
}

func NewOriginChecker() *OriginChecker {
	return &OriginChecker{
		allowedOrigins: make(map[string]bool),
		allowedASites:  make(map[string]string),
	}
}

func (oc *OriginChecker) AddOrigin(domain string) {
	oc.mu.Lock()
	defer oc.mu.Unlock()
	oc.allowedOrigins[domain] = true
}

func (oc *OriginChecker) RemoveOrigin(domain string) {
	oc.mu.Lock()
	defer oc.mu.Unlock()
	delete(oc.allowedOrigins, domain)
}

func (oc *OriginChecker) AddASite(domain, merchantID string) {
	oc.mu.Lock()
	defer oc.mu.Unlock()
	oc.allowedASites[domain] = merchantID
}

func (oc *OriginChecker) IsAllowed(origin string) bool {
	oc.mu.RLock()
	defer oc.mu.RUnlock()

	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := strings.ToLower(strings.TrimPrefix(u.Host, "www."))

	if oc.allowedOrigins[host] {
		return true
	}
	for domain := range oc.allowedOrigins {
		if strings.HasSuffix(host, "."+domain) {
			return true
		}
	}
	return false
}

func (oc *OriginChecker) GetMerchantForASite(domain string) (string, bool) {
	oc.mu.RLock()
	defer oc.mu.RUnlock()
	mid, ok := oc.allowedASites[domain]
	return mid, ok
}

func SignHMAC(key, message string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifyHMAC(key, message, signature string) bool {
	expected := SignHMAC(key, message)
	return hmac.Equal([]byte(expected), []byte(signature))
}
