package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/ab-payment-system/internal/admin/middleware"
)

// BSiteHandler handles full B-site CRUD management.
type BSiteHandler struct {
	db *sql.DB
}

// NewBSiteHandler creates a new handler.
func NewBSiteHandler(db *sql.DB) *BSiteHandler {
	return &BSiteHandler{db: db}
}

type BSiteRow struct {
	ID            string `json:"id"`
	Domain        string `json:"domain"`
	Name          string `json:"name"`
	HostingIP     string `json:"hosting_ip"`
	Status        string `json:"status"`
	WooURL        string `json:"woocommerce_url"`
	SSLExpires    string `json:"ssl_expires_at"`
	CreatedAt     string `json:"created_at"`
}

type BSiteDetail struct {
	ID            string `json:"id"`
	Domain        string `json:"domain"`
	Name          string `json:"name"`
	HostingIP     string `json:"hosting_ip"`
	HostingProv   string `json:"hosting_provider"`
	WooURL        string `json:"woocommerce_url"`
	WooKey        string `json:"woocommerce_key"`
	WooSecret     string `json:"woocommerce_secret"`
	SSLProvider   string `json:"ssl_provider"`
	SSLExpires    string `json:"ssl_expires_at"`
	CompanyInfo   string `json:"company_info"`
	BankAccounts  string `json:"bank_accounts"`
	SocialMedia   string `json:"social_media"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// List handles GET /api/v1/admin/b-sites
func (h *BSiteHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, domain, name, hosting_ip, status, woocommerce_url, ssl_expires_at, created_at
		 FROM b_sites ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"b_sites":[],"total":0}`))
		return
	}
	defer rows.Close()

	var sites []BSiteRow
	for rows.Next() {
		var s BSiteRow
		var ip, woo, ssl sql.NullString
		rows.Scan(&s.ID, &s.Domain, &s.Name, &ip, &s.Status, &woo, &ssl, &s.CreatedAt)
		if ip.Valid { s.HostingIP = ip.String }
		if woo.Valid { s.WooURL = woo.String }
		if ssl.Valid { s.SSLExpires = ssl.String }
		sites = append(sites, s)
	}
	if sites == nil { sites = []BSiteRow{} }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"b_sites": sites, "total": len(sites)})
}

// Create handles POST /api/v1/admin/b-sites
func (h *BSiteHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Domain        string `json:"domain"`
		Name          string `json:"name"`
		HostingIP     string `json:"hosting_ip"`
		HostingProv   string `json:"hosting_provider"`
		WooURL        string `json:"woocommerce_url"`
		WooKey        string `json:"woocommerce_key"`
		WooSecret     string `json:"woocommerce_secret"`
		SSLProvider   string `json:"ssl_provider"`
		SSLExpiresAt  string `json:"ssl_expires_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Domain == "" {
		http.Error(w, `{"error":"domain is required"}`, http.StatusBadRequest)
		return
	}

	adminID := middleware.GetAdminID(r.Context())

	var id, createdAt string
	err := h.db.QueryRowContext(r.Context(),
		`INSERT INTO b_sites (domain, name, hosting_ip, hosting_provider, woocommerce_url, woocommerce_key, woocommerce_secret, ssl_provider, ssl_expires_at, status)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'active')
		 RETURNING id, created_at`,
		req.Domain, req.Name,
		sql.NullString{String: req.HostingIP, Valid: req.HostingIP != ""},
		sql.NullString{String: req.HostingProv, Valid: req.HostingProv != ""},
		sql.NullString{String: req.WooURL, Valid: req.WooURL != ""},
		sql.NullString{String: req.WooKey, Valid: req.WooKey != ""},
		sql.NullString{String: req.WooSecret, Valid: req.WooSecret != ""},
		sql.NullString{String: req.SSLProvider, Valid: req.SSLProvider != ""},
		sql.NullString{String: req.SSLExpiresAt, Valid: req.SSLExpiresAt != ""},
	).Scan(&id, &createdAt)
	if err != nil {
		log.Error().Err(err).Str("admin_id", adminID).Str("domain", req.Domain).Msg("failed to create b-site")
		http.Error(w, `{"error":"create failed: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	log.Info().Str("admin_id", adminID).Str("b_site_id", id).Str("domain", req.Domain).Msg("b-site created")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": id, "domain": req.Domain, "name": req.Name, "status": "active", "created_at": createdAt,
	})
}

// GetByID handles GET /api/v1/admin/b-sites/{id}
func (h *BSiteHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var s BSiteDetail
	var ip, prov, woo, wook, woos, sslp, ssle, company, banks, social sql.NullString
	err := h.db.QueryRowContext(r.Context(),
		`SELECT id, domain, name, hosting_ip, hosting_provider, woocommerce_url, woocommerce_key, woocommerce_secret,
		 ssl_provider, ssl_expires_at, company_info, bank_accounts, social_media, status, created_at, updated_at
		 FROM b_sites WHERE id=$1`, id,
	).Scan(&s.ID, &s.Domain, &s.Name, &ip, &prov, &woo, &wook, &woos,
		&sslp, &ssle, &company, &banks, &social, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	if ip.Valid { s.HostingIP = ip.String }
	if prov.Valid { s.HostingProv = prov.String }
	if woo.Valid { s.WooURL = woo.String }
	if wook.Valid { s.WooKey = wook.String }
	if woos.Valid { s.WooSecret = woos.String }
	if sslp.Valid { s.SSLProvider = sslp.String }
	if ssle.Valid { s.SSLExpires = ssle.String }
	if company.Valid { s.CompanyInfo = company.String }
	if banks.Valid { s.BankAccounts = banks.String }
	if social.Valid { s.SocialMedia = social.String }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

// Update handles PUT /api/v1/admin/b-sites/{id}
func (h *BSiteHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var req struct {
		Name        string `json:"name"`
		Domain      string `json:"domain"`
		HostingIP   string `json:"hosting_ip"`
		Status      string `json:"status"`
		WooURL      string `json:"woocommerce_url"`
		WooKey      string `json:"woocommerce_key"`
		WooSecret   string `json:"woocommerce_secret"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	adminID := middleware.GetAdminID(r.Context())
	setClauses := "SET updated_at = NOW()"
	args := []interface{}{}
	argIdx := 1

	if req.Name != "" { setClauses += ", name=$" + itoa(argIdx); args = append(args, req.Name); argIdx++ }
	if req.Domain != "" { setClauses += ", domain=$" + itoa(argIdx); args = append(args, req.Domain); argIdx++ }
	if req.HostingIP != "" { setClauses += ", hosting_ip=$" + itoa(argIdx); args = append(args, req.HostingIP); argIdx++ }
	if req.Status != "" { setClauses += ", status=$" + itoa(argIdx); args = append(args, req.Status); argIdx++ }
	if req.WooURL != "" { setClauses += ", woocommerce_url=$" + itoa(argIdx); args = append(args, req.WooURL); argIdx++ }
	if req.WooKey != "" { setClauses += ", woocommerce_key=$" + itoa(argIdx); args = append(args, req.WooKey); argIdx++ }
	if req.WooSecret != "" { setClauses += ", woocommerce_secret=$" + itoa(argIdx); args = append(args, req.WooSecret); argIdx++ }

	if len(args) == 0 {
		http.Error(w, `{"error":"nothing to update"}`, http.StatusBadRequest)
		return
	}

	args = append(args, id)
	_, err := h.db.ExecContext(r.Context(),
		"UPDATE b_sites "+setClauses+" WHERE id=$"+itoa(argIdx), args...)
	if err != nil {
		log.Error().Err(err).Str("admin_id", adminID).Str("b_site_id", id).Msg("failed to update b-site")
		http.Error(w, `{"error":"update failed"}`, http.StatusInternalServerError)
		return
	}

	log.Info().Str("admin_id", adminID).Str("b_site_id", id).Msg("b-site updated")
	w.Write([]byte(`{"success":true}`))
}

// Delete handles DELETE /api/v1/admin/b-sites/{id}
func (h *BSiteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	adminID := middleware.GetAdminID(r.Context())
	if _, err := h.db.ExecContext(r.Context(), `DELETE FROM b_sites WHERE id=$1`, id); err != nil {
		http.Error(w, `{"error":"delete failed"}`, http.StatusInternalServerError)
		return
	}
	log.Info().Str("admin_id", adminID).Str("b_site_id", id).Msg("b-site deleted")
	w.Write([]byte(`{"success":true}`))
}

// UpdateStatus handles PATCH /api/v1/admin/b-sites/{id}/status
func (h *BSiteHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var req struct{ Status string `json:"status"` }
	json.NewDecoder(r.Body).Decode(&req)
	if req.Status == "" {
		http.Error(w, `{"error":"status is required"}`, http.StatusBadRequest)
		return
	}
	adminID := middleware.GetAdminID(r.Context())
	_, err := h.db.ExecContext(r.Context(),
		`UPDATE b_sites SET status=$1, updated_at=NOW() WHERE id=$2`, req.Status, id)
	if err != nil {
		http.Error(w, `{"error":"update failed"}`, http.StatusInternalServerError)
		return
	}
	log.Info().Str("admin_id", adminID).Str("b_site_id", id).Str("status", req.Status).Msg("b-site status updated")
	w.Write([]byte(`{"success":true}`))
}

func itoa(i int) string { return fmt.Sprintf("%d", i) }
