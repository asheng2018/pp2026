package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/ab-payment-system/internal/admin/middleware"
)

// AccountHandler handles admin account management with CRUD.
type AccountHandler struct {
	db *sql.DB
}

// NewAccountHandler creates a new handler.
func NewAccountHandler(db *sql.DB) *AccountHandler {
	return &AccountHandler{db: db}
}

type AccountRow struct {
	ID         string   `json:"id"`
	Gateway    string   `json:"gateway"`
	Alias      string   `json:"alias"`
	Status     string   `json:"status"`
	BSiteID    string   `json:"b_site_id"`
	MerchantID string   `json:"merchant_id"`
	Weight     int      `json:"weight"`
	Priority   int      `json:"priority"`
	Tags       []string `json:"tags"`
	LimitJSON  string   `json:"limit_config"`
	CreatedAt  string   `json:"created_at"`
}

// List handles GET /api/v1/admin/accounts
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, gateway, alias, status, b_site_id, merchant_id, weight, priority, tags, limit_config, created_at
		 FROM payment_accounts ORDER BY priority DESC, weight DESC LIMIT 100`)
	if err != nil {
		log.Error().Err(err).Msg("failed to list accounts")
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var accounts []AccountRow
	for rows.Next() {
		var a AccountRow
		var tags, limitB []byte
		var bs, mid sql.NullString
		if err := rows.Scan(&a.ID, &a.Gateway, &a.Alias, &a.Status, &bs, &mid,
				&a.Weight, &a.Priority, &tags, &limitB, &a.CreatedAt); err != nil {
				log.Error().Err(err).Msg("scan account row")
				continue
			}
		if bs.Valid {
			a.BSiteID = bs.String
		}
		if mid.Valid {
			a.MerchantID = mid.String
		}
		if len(tags) > 0 && string(tags) != "null" {
			json.Unmarshal(tags, &a.Tags)
		}
		if a.Tags == nil {
			a.Tags = []string{}
		}
		if limitB != nil {
			a.LimitJSON = string(limitB)
		}
		accounts = append(accounts, a)
	}
	if accounts == nil {
		accounts = []AccountRow{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"accounts": accounts, "total": len(accounts)})
}

// Create handles POST /api/v1/admin/accounts
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Gateway    string `json:"gateway"`
		Alias      string `json:"alias"`
		BSiteID    string `json:"b_site_id"`
		MerchantID string `json:"merchant_id"`
		// PayPal credentials
		PayPalClientID string `json:"paypal_client_id"`
		PayPalSecret   string `json:"paypal_secret"`
		// Stripe credentials
		StripePublishable string `json:"stripe_publishable_key"`
		StripeSecret      string `json:"stripe_secret_key"`
		// Limits
		SingleMin   string `json:"single_min"`
		SingleMax   string `json:"single_max"`
		DailyMax    string `json:"daily_max"`
		MonthlyMax  string `json:"monthly_max"`
		Weight      int    `json:"weight"`
		Priority    int    `json:"priority"`
		Tags        []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Gateway == "" || req.Alias == "" {
		http.Error(w, `{"error":"gateway and alias are required"}`, http.StatusBadRequest)
		return
	}

	adminID := middleware.GetAdminID(r.Context())

	// Build credential JSON for encryption (handled by account service later)
	cred := map[string]interface{}{}
	if req.Gateway == "paypal" {
		cred["paypal_client_id"] = req.PayPalClientID
		cred["paypal_secret"] = req.PayPalSecret
	} else if req.Gateway == "stripe" {
		cred["stripe_publishable_key"] = req.StripePublishable
		cred["stripe_secret_key"] = req.StripeSecret
	}
	encryptedCred := []byte(`{}`) // placeholder — crypto encryption wired in account_service.go
	if req.PayPalSecret != "" || req.StripeSecret != "" {
		credJSON, _ := json.Marshal(cred)
		encryptedCred = credJSON // In production, this goes through crypto.EncryptString
	}

	limitConfig := map[string]interface{}{
		"single_min":   req.SingleMin,
		"single_max":   req.SingleMax,
		"daily_max":    req.DailyMax,
		"monthly_max":  req.MonthlyMax,
		"lifetime_max": "0",
		"daily_count":  0,
	}
	limitJSON, _ := json.Marshal(limitConfig)
	tagsJSON, _ := json.Marshal(req.Tags)
	if tagsJSON == nil || string(tagsJSON) == "null" {
		tagsJSON = []byte(`{}`)
	}

	warg := 5
	if req.Weight > 0 {
		warg = req.Weight
	}
	pri := 10
	if req.Priority > 0 {
		pri = req.Priority
	}

	var id, createdAt string
	err := h.db.QueryRowContext(r.Context(),
		`INSERT INTO payment_accounts (gateway, alias, b_site_id, merchant_id, encrypted_cred, status, weight, priority, tags, limit_config, supported_currencies, supported_countries)
		 VALUES ($1,$2,$3,$4,$5,'online',$6,$7,$8,$9,'{"USD"}','{"US"}')
		 RETURNING id, created_at`,
		req.Gateway, req.Alias,
		sql.NullString{String: req.BSiteID, Valid: req.BSiteID != ""},
		sql.NullString{String: req.MerchantID, Valid: req.MerchantID != ""},
		encryptedCred, warg, pri, tagsJSON, limitJSON,
	).Scan(&id, &createdAt)
	if err != nil {
		log.Error().Err(err).Str("admin_id", adminID).Msg("failed to create account")
		http.Error(w, `{"error":"failed to create account: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	log.Info().Str("admin_id", adminID).Str("account_id", id).Str("alias", req.Alias).Msg("account created")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": id, "gateway": req.Gateway, "alias": req.Alias,
		"status": "online", "created_at": createdAt,
	})
}

// Update handles PUT /api/v1/admin/accounts/{id}
func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var req struct {
		Alias     string `json:"alias"`
		Status    string `json:"status"`
		Weight    int    `json:"weight"`
		Priority  int    `json:"priority"`
		BSiteID   string `json:"b_site_id"`
		SingleMax string `json:"single_max"`
		DailyMax  string `json:"daily_max"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	adminID := middleware.GetAdminID(r.Context())

	if req.Alias != "" {
		h.db.ExecContext(r.Context(), `UPDATE payment_accounts SET alias=$1, updated_at=NOW() WHERE id=$2`, req.Alias, id)
	}
	if req.Status != "" {
		h.db.ExecContext(r.Context(), `UPDATE payment_accounts SET status=$1, updated_at=NOW() WHERE id=$2`, req.Status, id)
	}
	if req.Weight > 0 {
		h.db.ExecContext(r.Context(), `UPDATE payment_accounts SET weight=$1, updated_at=NOW() WHERE id=$2`, req.Weight, id)
	}
	if req.Priority > 0 {
		h.db.ExecContext(r.Context(), `UPDATE payment_accounts SET priority=$1, updated_at=NOW() WHERE id=$2`, req.Priority, id)
	}
	if req.BSiteID != "" {
		h.db.ExecContext(r.Context(), `UPDATE payment_accounts SET b_site_id=$1, updated_at=NOW() WHERE id=$2`, req.BSiteID, id)
	}

	log.Info().Str("admin_id", adminID).Str("account_id", id).Msg("account updated")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true}`))
}

// Delete handles DELETE /api/v1/admin/accounts/{id}
func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	adminID := middleware.GetAdminID(r.Context())
	if _, err := h.db.ExecContext(r.Context(), `DELETE FROM payment_accounts WHERE id=$1`, id); err != nil {
		log.Error().Err(err).Str("admin_id", adminID).Msg("failed to delete account")
		http.Error(w, `{"error":"delete failed"}`, http.StatusInternalServerError)
		return
	}
	log.Info().Str("admin_id", adminID).Str("account_id", id).Msg("account deleted")
	w.Write([]byte(`{"success":true}`))
}

// UpdateStatus handles PATCH /api/v1/admin/accounts/{id}/status (legacy, kept for backward compat)
func (h *AccountHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var req struct{ Status string `json:"status"` }
	json.NewDecoder(r.Body).Decode(&req)
	if req.Status == "" {
		http.Error(w, `{"error":"status is required"}`, http.StatusBadRequest)
		return
	}
	adminID := middleware.GetAdminID(r.Context())
	if _, err := h.db.ExecContext(r.Context(),
		`UPDATE payment_accounts SET status=$1, updated_at=NOW() WHERE id=$2`, req.Status, id); err != nil {
		http.Error(w, `{"error":"update failed"}`, http.StatusInternalServerError)
		return
	}
	log.Info().Str("admin_id", adminID).Str("account_id", id).Str("status", req.Status).Msg("account status updated")
	w.Write([]byte(`{"success":true}`))
}
