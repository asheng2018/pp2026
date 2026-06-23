package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// ExchangeRateListHandler handles exchange rate listing.
type ExchangeRateHandler struct {
	db *sql.DB
}

// NewExchangeRateHandler creates a new handler.
func NewExchangeRateHandler(db *sql.DB) *ExchangeRateHandler {
	return &ExchangeRateHandler{db: db}
}

type ExchangeRateRow struct {
	ID             string `json:"id"`
	BaseCurrency   string `json:"base_currency"`
	TargetCurrency string `json:"target_currency"`
	Rate           string `json:"rate"`
	Source         string `json:"source"`
	FetchedAt      string `json:"fetched_at"`
}

// List handles GET /api/v1/admin/exchange-rates
func (h *ExchangeRateHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, base_currency, target_currency, rate, source, fetched_at
		 FROM exchange_rates ORDER BY fetched_at DESC LIMIT 100`)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"rates":[],"total":0}`))
		return
	}
	defer rows.Close()

	var rates []ExchangeRateRow
	for rows.Next() {
		var e ExchangeRateRow
		var src sql.NullString
		rows.Scan(&e.ID, &e.BaseCurrency, &e.TargetCurrency, &e.Rate, &src, &e.FetchedAt)
		if src.Valid { e.Source = src.String }
		rates = append(rates, e)
	}
	if rates == nil { rates = []ExchangeRateRow{} }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"rates": rates, "total": len(rates)})
}
