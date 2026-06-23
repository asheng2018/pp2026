package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/ab-payment-system/internal/admin/service"
)

type contextKey string

const (
	AdminIDContextKey   contextKey = "admin_id"
	AdminRoleContextKey contextKey = "admin_role"
)

// Middleware returns a gorilla/mux-compatible middleware that validates
// admin JWT tokens and injects admin identity into the request context.
func Middleware(tokenSvc *service.TokenService, blacklist *service.TokenBlacklist) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractBearerToken(r)
			if tokenStr == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			claims, err := tokenSvc.Validate(tokenStr)
			if err != nil {
				log.Warn().Err(err).Msg("admin jwt validation failed")
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			// Check blacklist
			if blacklisted, err := blacklist.IsBlacklisted(r.Context(), claims.ID); err != nil {
				log.Error().Err(err).Msg("failed to check token blacklist")
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
				return
			} else if blacklisted {
				http.Error(w, `{"error":"token has been revoked"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), AdminIDContextKey, claims.AdminID)
			ctx = context.WithValue(ctx, AdminRoleContextKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAdminID extracts the admin UUID from the request context.
// Returns an empty string if not present.
func GetAdminID(ctx context.Context) string {
	if id, ok := ctx.Value(AdminIDContextKey).(string); ok {
		return id
	}
	return ""
}

// GetAdminRole extracts the admin role from the request context.
// Returns an empty string if not present.
func GetAdminRole(ctx context.Context) string {
	if role, ok := ctx.Value(AdminRoleContextKey).(string); ok {
		return role
	}
	return ""
}

// extractBearerToken pulls a Bearer token from the Authorization header.
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}
