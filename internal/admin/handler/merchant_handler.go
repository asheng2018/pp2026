package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"github.com/ab-payment-system/internal/admin/middleware"
	"github.com/ab-payment-system/internal/merchant/model"
	merchantRepo "github.com/ab-payment-system/internal/merchant/repository"
)

// MerchantHandler handles full merchant CRUD + API key management.
type MerchantHandler struct {
	repo     *merchantRepo.MerchantRepo
	jwtSecret string
}

// NewMerchantHandler creates a new handler.
func NewMerchantHandler(repo *merchantRepo.MerchantRepo, jwtSecret string) *MerchantHandler {
	return &MerchantHandler{repo: repo, jwtSecret: jwtSecret}
}

// Create handles POST /api/v1/admin/merchants
func (h *MerchantHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string  `json:"name"`
		Email    string  `json:"email"`
		FeeRate  float64 `json:"fee_rate"`
		Routing  string  `json:"routing_mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}

	adminID := middleware.GetAdminID(r.Context())

	m := &model.Merchant{
		Name:        req.Name,
		Email:       req.Email,
		Status:      "active",
		RoutingMode: req.Routing,
		RiskProfile: &model.RiskProfile{MaxOrderAmount: 5000, CountryBlacklist: []string{}},
		FeeConfig:   &model.FeeConfig{
			PayPalFee: &model.GatewayFee{FixedFee: decimal.NewFromFloat(0.30), RateFee: decimal.NewFromFloat(0.044), MinFee: decimal.NewFromFloat(0.50)},
			StripeFee: &model.GatewayFee{FixedFee: decimal.NewFromFloat(0.30), RateFee: decimal.NewFromFloat(0.029), MinFee: decimal.NewFromFloat(0.50)},
		},
		Settlement:  &model.SettlementConfig{Method: "wire", MinAmount: decimal.NewFromFloat(100), Cycle: "weekly"},
		Metadata:    map[string]interface{}{"created_by": adminID},
	}

	if err := h.repo.Create(r.Context(), m); err != nil {
		log.Error().Err(err).Str("admin_id", adminID).Msg("failed to create merchant")
		http.Error(w, `{"error":"failed to create merchant"}`, http.StatusInternalServerError)
		return
	}

	log.Info().Str("admin_id", adminID).Str("merchant_id", m.ID).Str("name", m.Name).Msg("merchant created")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

	// List handles GET /api/v1/admin/merchants
	func (h *MerchantHandler) List(w http.ResponseWriter, r *http.Request) {
		merchants, err := h.repo.List(r.Context(), 0, 100)
		if err != nil {
			log.Error().Err(err).Msg("failed to list merchants")
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"merchants":[],"total":0}`))
			return
		}
		if merchants == nil {
			merchants = []*model.Merchant{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"merchants": merchants, "total": len(merchants)})
	}


// GetByID handles GET /api/v1/admin/merchants/{id}
func (h *MerchantHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	m, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"merchant not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// Update handles PUT /api/v1/admin/merchants/{id}
func (h *MerchantHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	m, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"merchant not found"}`, http.StatusNotFound)
		return
	}

	var req struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Status string `json:"status"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Name != "" {
		m.Name = req.Name
	}
	if req.Email != "" {
		m.Email = req.Email
	}
	if req.Status != "" {
		m.Status = req.Status
	}

	if err := h.repo.Update(r.Context(), m); err != nil {
		log.Error().Err(err).Msg("failed to update merchant")
		http.Error(w, `{"error":"update failed"}`, http.StatusInternalServerError)
		return
	}

	adminID := middleware.GetAdminID(r.Context())
	log.Info().Str("admin_id", adminID).Str("merchant_id", id).Msg("merchant updated")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// Delete handles DELETE /api/v1/admin/merchants/{id}
func (h *MerchantHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.repo.Delete(r.Context(), id); err != nil {
		http.Error(w, `{"error":"delete failed"}`, http.StatusInternalServerError)
		return
	}
	adminID := middleware.GetAdminID(r.Context())
	log.Info().Str("admin_id", adminID).Str("merchant_id", id).Msg("merchant deleted")
	w.Write([]byte(`{"success":true}`))
}

// GenerateAPIKey handles POST /api/v1/admin/merchants/{id}/apikeys
func (h *MerchantHandler) GenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	merchantID := mux.Vars(r)["id"]

	var req struct {
		Permissions  []string `json:"permissions"`
		IPWhitelist  []string `json:"ip_whitelist"`
		ExpiresDays  int      `json:"expires_days"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if len(req.Permissions) == 0 {
		req.Permissions = []string{"read", "write"}
	}

	rawKey := generateSecureToken(40)
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	k := &model.APIKey{
		KeyHash:     keyHash,
		KeyPrefix:   rawKey[:8],
		Permissions: req.Permissions,
		IPWhitelist: req.IPWhitelist,
	}

	if err := h.repo.CreateAPIKey(r.Context(), merchantID, k); err != nil {
		log.Error().Err(err).Str("merchant_id", merchantID).Msg("failed to create API key")
		http.Error(w, `{"error":"failed to create API key"}`, http.StatusInternalServerError)
		return
	}

	adminID := middleware.GetAdminID(r.Context())
	log.Info().Str("admin_id", adminID).Str("merchant_id", merchantID).Str("key_id", k.KeyID).Msg("API key generated")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"key_id":      k.KeyID,
		"api_key":     rawKey,
		"key_prefix":  k.KeyPrefix,
		"permissions": k.Permissions,
		"message":     "Save this API key now — it will not be shown again.",
	})
}

// ListAPIKeys handles GET /api/v1/admin/merchants/{id}/apikeys
func (h *MerchantHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	merchantID := mux.Vars(r)["id"]
	keys, err := h.repo.ListAPIKeys(r.Context(), merchantID)
	if err != nil {
		http.Error(w, `{"error":"failed to list keys"}`, http.StatusInternalServerError)
		return
	}
	if keys == nil {
		keys = []*model.APIKey{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"keys": keys, "total": len(keys)})
}

// RevokeAPIKey handles DELETE /api/v1/admin/merchants/{id}/apikeys/{keyId}
func (h *MerchantHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	keyID := mux.Vars(r)["keyId"]
	if err := h.repo.RevokeAPIKey(r.Context(), keyID); err != nil {
		http.Error(w, `{"error":"revoke failed"}`, http.StatusInternalServerError)
		return
	}
	adminID := middleware.GetAdminID(r.Context())
	log.Info().Str("admin_id", adminID).Str("key_id", keyID).Msg("API key revoked")
	w.Write([]byte(`{"success":true}`))
}

func generateSecureToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)[:length]
}
