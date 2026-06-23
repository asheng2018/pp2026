package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/ab-payment-system/internal/merchant/model"
)

type Authenticator interface {
	AuthenticateAPIKey(ctx context.Context, rawKey string) (*model.Merchant, *model.APIKey, error)
}

type contextKey string

const (
	MerchantContextKey contextKey = "merchant"
	APIKeyContextKey   contextKey = "api_key"
)

func Middleware(auth Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawKey := extractAPIKey(r)
			if rawKey == "" {
				http.Error(w, `{"error":"missing api key"}`, http.StatusUnauthorized)
				return
			}

			merchant, key, err := auth.AuthenticateAPIKey(r.Context(), rawKey)
			if err != nil {
				log.Warn().Err(err).Str("key_prefix", rawKey[:min(10, len(rawKey))]).Msg("api key auth failed")
				http.Error(w, `{"error":"invalid api key"}`, http.StatusUnauthorized)
				return
			}

			// Check IP whitelist
			if len(key.IPWhitelist) > 0 {
				clientIP := extractClientIP(r)
				if !isIPAllowed(clientIP, key.IPWhitelist) {
					log.Warn().Str("merchant", merchant.ID).Str("ip", clientIP).Msg("ip not in whitelist")
					http.Error(w, `{"error":"ip not allowed"}`, http.StatusForbidden)
					return
				}
			}

			ctx := context.WithValue(r.Context(), MerchantContextKey, merchant)
			ctx = context.WithValue(ctx, APIKeyContextKey, key)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetMerchant(ctx context.Context) *model.Merchant {
	if m, ok := ctx.Value(MerchantContextKey).(*model.Merchant); ok {
		return m
	}
	return nil
}

func GetAPIKey(ctx context.Context) *model.APIKey {
	if k, ok := ctx.Value(APIKeyContextKey).(*model.APIKey); ok {
		return k
	}
	return nil
}

func extractAPIKey(r *http.Request) string {
	// Check Authorization header first
	if auth := r.Header.Get("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
	}
	// Check X-API-Key header
	if key := r.Header.Get("X-API-Key"); key != "" {
		return key
	}
	// Check query parameter
	if key := r.URL.Query().Get("api_key"); key != "" {
		return key
	}
	return ""
}

func extractClientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.TrimSpace(strings.Split(fwd, ",")[0])
	}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}

func isIPAllowed(clientIP string, whitelist []string) bool {
	for _, allowed := range whitelist {
		if allowed == clientIP || allowed == "0.0.0.0/0" {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
