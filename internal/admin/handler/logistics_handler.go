package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// LogisticsListHandler handles logistics tracking listing.
type LogisticsHandler struct {
	db *sql.DB
}

// NewLogisticsHandler creates a new handler.
func NewLogisticsHandler(db *sql.DB) *LogisticsHandler {
	return &LogisticsHandler{db: db}
}

type LogisticsRow struct {
	ID              string `json:"id"`
	OrderID         string `json:"order_id"`
	TrackingNumber  string `json:"tracking_number"`
	Carrier         string `json:"carrier"`
	Status          string `json:"status"`
	SyncedToBSite   bool   `json:"synced_to_b_site"`
	CreatedAt       string `json:"created_at"`
}

// List handles GET /api/v1/admin/logistics
func (h *LogisticsHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, order_id, tracking_number, carrier, status, synced_to_b_site, created_at
		 FROM logistics_tracking ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tracks":[],"total":0}`))
		return
	}
	defer rows.Close()

	var tracks []LogisticsRow
	for rows.Next() {
		var t LogisticsRow
		var oid, tn, carrier sql.NullString
		rows.Scan(&t.ID, &oid, &tn, &carrier, &t.Status, &t.SyncedToBSite, &t.CreatedAt)
		if oid.Valid { t.OrderID = oid.String }
		if tn.Valid { t.TrackingNumber = tn.String }
		if carrier.Valid { t.Carrier = carrier.String }
		tracks = append(tracks, t)
	}
	if tracks == nil { tracks = []LogisticsRow{} }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"tracks": tracks, "total": len(tracks)})
}
