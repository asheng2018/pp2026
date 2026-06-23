package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// SettlementListHandler handles settlement management.
type SettlementListHandler struct {
	db *sql.DB
}

// NewSettlementListHandler creates a new handler.
func NewSettlementListHandler(db *sql.DB) *SettlementListHandler {
	return &SettlementListHandler{db: db}
}

type SettlementRow struct {
	ID           string `json:"id"`
	MerchantID   string `json:"merchant_id"`
	CycleStart   string `json:"cycle_start"`
	CycleEnd     string `json:"cycle_end"`
	TotalOrders  int    `json:"total_orders"`
	TotalAmount  string `json:"total_amount"`
	TotalFee     string `json:"total_fee"`
	NetAmount    string `json:"net_amount"`
	PayoutMethod string `json:"payout_method"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
}

// List handles GET /api/v1/admin/settlements
func (h *SettlementListHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, merchant_id, cycle_start, cycle_end, total_orders, total_amount, total_fee, net_amount,
		 payout_method, status, created_at FROM settlements ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		log.Error().Err(err).Msg("failed to list settlements")
		// settlements table may be empty
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"settlements":[],"total":0}`))
		return
	}
	defer rows.Close()

	var settlements []SettlementRow
	for rows.Next() {
		var s SettlementRow
		var pm sql.NullString
		rows.Scan(&s.ID, &s.MerchantID, &s.CycleStart, &s.CycleEnd, &s.TotalOrders,
			&s.TotalAmount, &s.TotalFee, &s.NetAmount, &pm, &s.Status, &s.CreatedAt)
		if pm.Valid { s.PayoutMethod = pm.String }
		settlements = append(settlements, s)
	}
	if settlements == nil { settlements = []SettlementRow{} }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"settlements": settlements, "total": len(settlements)})
}
