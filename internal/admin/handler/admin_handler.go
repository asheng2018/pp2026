package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/ab-payment-system/internal/admin/middleware"
	"github.com/ab-payment-system/internal/admin/service"
)

// AdminHandler handles HTTP requests for admin authentication.
type AdminHandler struct {
	svc *service.AdminService
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(svc *service.AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

// Login handles POST /api/v1/admin/login.
func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, `{"error":"username and password are required"}`, http.StatusBadRequest)
		return
	}

	result, err := h.svc.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			log.Warn().Str("username", req.Username).Msg("admin login failed: invalid credentials")
			http.Error(w, `{"error":"invalid username or password"}`, http.StatusUnauthorized)
			return
		}
		log.Error().Err(err).Str("username", req.Username).Msg("admin login error")
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	log.Info().Str("username", result.Username).Str("role", result.Role).Msg("admin logged in")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Logout handles POST /api/v1/admin/logout.
func (h *AdminHandler) Logout(w http.ResponseWriter, r *http.Request) {
	tokenStr := extractTokenFromRequest(r)
	if tokenStr == "" {
		http.Error(w, `{"error":"missing authorization header"}`, http.StatusBadRequest)
		return
	}

	if err := h.svc.Logout(r.Context(), tokenStr); err != nil {
		log.Warn().Err(err).Msg("admin logout failed")
		// Still return success to the client — the token is invalid anyway.
	}

	adminID := middleware.GetAdminID(r.Context())
	log.Info().Str("admin_id", adminID).Msg("admin logged out")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true}`))
}

// Me handles GET /api/v1/admin/me.
func (h *AdminHandler) Me(w http.ResponseWriter, r *http.Request) {
	tokenStr := extractTokenFromRequest(r)
	if tokenStr == "" {
		http.Error(w, `{"error":"missing authorization header"}`, http.StatusBadRequest)
		return
	}

	user, err := h.svc.GetCurrentUser(r.Context(), tokenStr)
	if err != nil {
		log.Warn().Err(err).Msg("admin /me lookup failed")
		http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// extractTokenFromRequest pulls the Bearer token from the Authorization header,
// or falls back to the context if the middleware already ran.
func extractTokenFromRequest(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return ""
}
