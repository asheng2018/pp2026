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

// ProxyHandler handles full proxy CRUD management.
type ProxyHandler struct {
	db *sql.DB
}

// NewProxyHandler creates a new handler.
func NewProxyHandler(db *sql.DB) *ProxyHandler {
	return &ProxyHandler{db: db}
}

type ProxyRow struct {
	ID          string  `json:"id"`
	Type        string  `json:"proxy_type"`
	Host        string  `json:"host"`
	Port        int     `json:"port"`
	Username    string  `json:"username"`
	Protocol    string  `json:"protocol"`
	Country     string  `json:"country"`
	City        string  `json:"city"`
	ISP         string  `json:"isp"`
	Status      string  `json:"status"`
	Latency     int     `json:"latency"`
	SuccessRate float64 `json:"success_rate"`
	BoundAcc    string  `json:"bound_account_id"`
	IsDedicated bool    `json:"is_dedicated"`
	CreatedAt   string  `json:"created_at"`
}

// List handles GET /api/v1/admin/proxies
func (h *ProxyHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, proxy_type, protocol, host, port, username, country, city, isp, status, latency,
		 success_rate, bound_account_id, is_dedicated, created_at
		 FROM proxies ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"proxies":[],"total":0}`))
		return
	}
	defer rows.Close()

	var proxies []ProxyRow
	for rows.Next() {
		var p ProxyRow
		var country, city, isp, username, bound sql.NullString
		rows.Scan(&p.ID, &p.Type, &p.Protocol, &p.Host, &p.Port, &username,
			&country, &city, &isp, &p.Status, &p.Latency, &p.SuccessRate, &bound, &p.IsDedicated, &p.CreatedAt)
		if country.Valid { p.Country = country.String }
		if city.Valid { p.City = city.String }
		if isp.Valid { p.ISP = isp.String }
		if username.Valid { p.Username = username.String }
		if bound.Valid { p.BoundAcc = bound.String }
		proxies = append(proxies, p)
	}
	if proxies == nil { proxies = []ProxyRow{} }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"proxies": proxies, "total": len(proxies)})
}

// Create handles POST /api/v1/admin/proxies
func (h *ProxyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type          string `json:"proxy_type"`
		Protocol      string `json:"protocol"`
		Host          string `json:"host"`
		Port          int    `json:"port"`
		Username      string `json:"username"`
		Password      string `json:"password"`
		Country       string `json:"country"`
		City          string `json:"city"`
		ISP           string `json:"isp"`
		BoundAccountID string `json:"bound_account_id"`
		IsDedicated   bool   `json:"is_dedicated"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Host == "" || req.Port == 0 {
		http.Error(w, `{"error":"host and port are required"}`, http.StatusBadRequest)
		return
	}
	if req.Type == "" { req.Type = "residential" }
	if req.Protocol == "" { req.Protocol = "socks5" }

	adminID := middleware.GetAdminID(r.Context())

	var id, createdAt string
	err := h.db.QueryRowContext(r.Context(),
		`INSERT INTO proxies (proxy_type, protocol, host, port, username, encrypted_password, country, city, isp, bound_account_id, is_dedicated, status)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'testing')
		 RETURNING id, created_at`,
		req.Type, req.Protocol, req.Host, req.Port,
		sql.NullString{String: req.Username, Valid: req.Username != ""},
		[]byte(req.Password),
		sql.NullString{String: req.Country, Valid: req.Country != ""},
		sql.NullString{String: req.City, Valid: req.City != ""},
		sql.NullString{String: req.ISP, Valid: req.ISP != ""},
		sql.NullString{String: req.BoundAccountID, Valid: req.BoundAccountID != ""},
		req.IsDedicated,
	).Scan(&id, &createdAt)
	if err != nil {
		log.Error().Err(err).Str("admin_id", adminID).Msg("failed to create proxy")
		http.Error(w, `{"error":"create failed: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	log.Info().Str("admin_id", adminID).Str("proxy_id", id).Str("host", req.Host).Msg("proxy created")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": id, "proxy_type": req.Type, "host": req.Host, "port": req.Port, "status": "testing", "created_at": createdAt,
	})
}

// BatchImport handles POST /api/v1/admin/proxies/batch
func (h *ProxyHandler) BatchImport(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Proxies []struct {
			Type     string `json:"proxy_type"`
			Proto    string `json:"protocol"`
			Host     string `json:"host"`
			Port     int    `json:"port"`
			User     string `json:"username"`
			Pass     string `json:"password"`
			Country  string `json:"country"`
			City     string `json:"city"`
			ISP      string `json:"isp"`
		} `json:"proxies"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	adminID := middleware.GetAdminID(r.Context())
	imported := 0
	for _, px := range req.Proxies {
		if px.Host == "" || px.Port == 0 { continue }
		if px.Type == "" { px.Type = "residential" }
		if px.Proto == "" { px.Proto = "socks5" }
		_, err := h.db.ExecContext(r.Context(),
			`INSERT INTO proxies (proxy_type, protocol, host, port, username, encrypted_password, country, city, isp, status)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'testing')
			 ON CONFLICT DO NOTHING`,
			px.Type, px.Proto, px.Host, px.Port,
			sql.NullString{String: px.User, Valid: px.User != ""},
			[]byte(px.Pass),
			sql.NullString{String: px.Country, Valid: px.Country != ""},
			sql.NullString{String: px.City, Valid: px.City != ""},
			sql.NullString{String: px.ISP, Valid: px.ISP != ""},
		)
		if err == nil { imported++ }
	}

	log.Info().Str("admin_id", adminID).Int("imported", imported).Int("total", len(req.Proxies)).Msg("proxies batch imported")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true, "imported": imported, "total": len(req.Proxies),
	})
}

// Update handles PUT /api/v1/admin/proxies/{id}
func (h *ProxyHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var req struct {
		Status string `json:"status"`
		Latency int   `json:"latency"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	adminID := middleware.GetAdminID(r.Context())

	if req.Status != "" {
		h.db.ExecContext(r.Context(), `UPDATE proxies SET status=$1, updated_at=NOW() WHERE id=$2`, req.Status, id)
	}
	if req.Latency > 0 {
		h.db.ExecContext(r.Context(), `UPDATE proxies SET latency=$1, updated_at=NOW() WHERE id=$2`, req.Latency, id)
	}

	log.Info().Str("admin_id", adminID).Str("proxy_id", id).Msg("proxy updated")
	w.Write([]byte(`{"success":true}`))
}

// Delete handles DELETE /api/v1/admin/proxies/{id}
func (h *ProxyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	adminID := middleware.GetAdminID(r.Context())
	if _, err := h.db.ExecContext(r.Context(), `DELETE FROM proxies WHERE id=$1`, id); err != nil {
		http.Error(w, `{"error":"delete failed"}`, http.StatusInternalServerError)
		return
	}
	log.Info().Str("admin_id", adminID).Str("proxy_id", id).Msg("proxy deleted")
	w.Write([]byte(`{"success":true}`))
}

// UpdateStatus handles PATCH /api/v1/admin/proxies/{id}/status
func (h *ProxyHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var req struct{ Status string `json:"status"` }
	json.NewDecoder(r.Body).Decode(&req)
	if req.Status == "" {
		http.Error(w, `{"error":"status is required"}`, http.StatusBadRequest)
		return
	}
	adminID := middleware.GetAdminID(r.Context())
	_, err := h.db.ExecContext(r.Context(),
		`UPDATE proxies SET status=$1, updated_at=NOW() WHERE id=$2`, req.Status, id)
	if err != nil {
		http.Error(w, `{"error":"update failed"}`, http.StatusInternalServerError)
		return
	}
	log.Info().Str("admin_id", adminID).Str("proxy_id", id).Str("status", req.Status).Msg("proxy status updated")
	w.Write([]byte(`{"success":true}`))
}

func pfx(format string, a ...interface{}) string { return fmt.Sprintf(format, a...) }
