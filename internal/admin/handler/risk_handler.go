package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// RiskListHandler handles risk event listing.
type RiskListHandler struct {
	db *sql.DB
}

// NewRiskListHandler creates a new handler.
func NewRiskListHandler(db *sql.DB) *RiskListHandler {
	return &RiskListHandler{db: db}
}

type RiskRow struct {
	ID         string `json:"id"`
	MerchantID string `json:"merchant_id"`
	OrderID    string `json:"order_id"`
	RuleName   string `json:"rule_name"`
	RiskLevel  string `json:"risk_level"`
	RiskScore  float64 `json:"risk_score"`
	Action     string `json:"action"`
	Reason     string `json:"reason"`
	CreatedAt  string `json:"created_at"`
}

// List handles GET /api/v1/admin/risk-events
func (h *RiskListHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, merchant_id, order_id, rule_name, risk_level, risk_score, action, reason, created_at
		 FROM risk_events ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"events":[],"total":0}`))
		return
	}
	defer rows.Close()

	var events []RiskRow
	for rows.Next() {
		var e RiskRow
		var oid, reason sql.NullString
		rows.Scan(&e.ID, &e.MerchantID, &oid, &e.RuleName, &e.RiskLevel, &e.RiskScore, &e.Action, &reason, &e.CreatedAt)
		if oid.Valid { e.OrderID = oid.String }
		if reason.Valid { e.Reason = reason.String }
		events = append(events, e)
	}
	if events == nil { events = []RiskRow{} }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"events": events, "total": len(events)})
}
