package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/ab-payment-system/internal/admin/middleware"
)

// DashboardHandler provides aggregated stats for the admin dashboard.
type DashboardHandler struct {
	db *sql.DB
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(db *sql.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

// DashboardStats represents the aggregated data displayed on the admin dashboard.
type DashboardStats struct {
	TotalOrders     int     `json:"total_orders"`
	TodayOrders     int     `json:"today_orders"`
	TodayRevenue    string  `json:"today_revenue"`
	ActiveAccounts  int     `json:"active_accounts"`
	ActiveMerchants int     `json:"active_merchants"`
	SuccessRate     float64 `json:"success_rate"`
	PendingOrders   int     `json:"pending_orders"`
	FailedOrders    int     `json:"failed_orders"`
}

// GetStats handles GET /api/v1/admin/dashboard/stats
func (h *DashboardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var stats DashboardStats

	// Total orders
	h.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM orders`).Scan(&stats.TotalOrders)

	// Today's orders
	h.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM orders WHERE created_at::date = CURRENT_DATE`).Scan(&stats.TodayOrders)

	// Today's revenue (paid orders)
	var revenue sql.NullString
	h.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(amount)::text, '0') FROM orders WHERE created_at::date = CURRENT_DATE AND status IN ('paid','completed')`).Scan(&revenue)
	if revenue.Valid {
		stats.TodayRevenue = revenue.String
	} else {
		stats.TodayRevenue = "0"
	}

	// Active accounts
	h.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM payment_accounts WHERE status = 'online'`).Scan(&stats.ActiveAccounts)

	// Active merchants
	h.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM merchants WHERE status = 'active'`).Scan(&stats.ActiveMerchants)

	// Success rate
	var total int64
	var paid int64
	h.db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(SUM(CASE WHEN status IN ('paid','completed') THEN 1 ELSE 0 END),0) FROM orders`).Scan(&total, &paid)
	if total > 0 {
		stats.SuccessRate = float64(paid) / float64(total) * 100
	} else {
		stats.SuccessRate = 100
	}

	// Pending orders
	h.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM orders WHERE status IN ('pending','processing')`).Scan(&stats.PendingOrders)

	// Failed orders
	h.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM orders WHERE status IN ('failed','canceled')`).Scan(&stats.FailedOrders)

	adminID := middleware.GetAdminID(ctx)
	log.Info().Str("admin_id", adminID).Msg("dashboard stats requested")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
