package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/ab-payment-system/internal/admin/middleware"
	"github.com/ab-payment-system/internal/order/model"
	orderRepo "github.com/ab-payment-system/internal/order/repository"
)

// OrderHandler handles admin order management requests.
type OrderHandler struct {
	repo *orderRepo.OrderRepo
	db   *sql.DB
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(repo *orderRepo.OrderRepo, db *sql.DB) *OrderHandler {
	return &OrderHandler{repo: repo, db: db}
}

// List handles GET /api/v1/admin/orders
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit := 50
	offset := 0

	adminID := middleware.GetAdminID(r.Context())

	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, order_no, merchant_id, account_id, gateway, amount, currency, status,
		 customer_email, customer_country, risk_level, risk_score, created_at, updated_at
		 FROM orders `+buildFilter(status)+` ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		log.Error().Err(err).Str("admin_id", adminID).Msg("failed to list orders")
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		o := &model.Order{}
		rows.Scan(&o.ID, &o.OrderNo, &o.MerchantID, &o.AccountID, &o.Gateway,
			&o.Amount, &o.Currency, &o.Status, &o.CustomerEmail, &o.CustomerCountry,
			&o.RiskLevel, &o.RiskScore, &o.CreatedAt, &o.UpdatedAt)
		orders = append(orders, o)
	}

	// Total count
	var total int
	var countQuery string
	var countArgs []interface{}
	if status != "" {
		countQuery = `SELECT COUNT(*) FROM orders WHERE status = $1`
		countArgs = []interface{}{status}
	} else {
		countQuery = `SELECT COUNT(*) FROM orders`
	}
	h.db.QueryRowContext(r.Context(), countQuery, countArgs...).Scan(&total)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"orders": orders,
		"total":  total,
	})
}

// GetByID handles GET /api/v1/admin/orders/{id}
func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	order, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"order not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// GetDailyRevenue handles GET /api/v1/admin/dashboard/revenue
func (h *OrderHandler) GetDailyRevenue(w http.ResponseWriter, r *http.Request) {
	days := 7
	type daily struct {
		Date    string `json:"date"`
		Orders  int    `json:"orders"`
		Revenue string `json:"revenue"`
	}
	var result []daily
	for i := days - 1; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		var count int
		var rev sql.NullString
		h.db.QueryRowContext(r.Context(),
			`SELECT COUNT(*), COALESCE(SUM(amount)::text, '0') FROM orders WHERE created_at::date = $1`, d,
		).Scan(&count, &rev)
		rStr := "0"
		if rev.Valid {
			rStr = rev.String
		}
		result = append(result, daily{Date: d, Orders: count, Revenue: rStr})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func buildFilter(status string) string {
	if status != "" {
		return "WHERE status = '" + status + "'"
	}
	return ""
}
