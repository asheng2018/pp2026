package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"

	"github.com/ab-payment-system/internal/order/model"
)

type OrderRepo struct{ db *sql.DB }

func NewOrderRepo(db *sql.DB) *OrderRepo { return &OrderRepo{db: db} }

func (r *OrderRepo) Create(ctx context.Context, o *model.Order) error {
	cb, _ := json.Marshal(o.CallbackData)
	md, _ := json.Marshal(o.Metadata)
	return r.db.QueryRowContext(ctx, `
		INSERT INTO orders (order_no, merchant_id, account_id, gateway, amount, currency,
			status, customer_email, customer_ip, customer_country, pay_token_hash,
			risk_level, risk_score, callback_data, a_site_referer, metadata, expired_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
		RETURNING id, created_at, updated_at`,
		o.OrderNo, o.MerchantID, o.AccountID, o.Gateway, o.Amount, o.Currency,
		o.Status, o.CustomerEmail, o.CustomerIP, o.CustomerCountry, o.PayTokenHash,
		o.RiskLevel, o.RiskScore, cb, o.ASiteReferer, md, o.ExpiredAt, time.Now(), time.Now(),
	).Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
}

func (r *OrderRepo) FindByID(ctx context.Context, id string) (*model.Order, error) {
	o := &model.Order{}
	var paidAt, expiredAt, canceledAt sql.NullTime
	var cb, md []byte
	err := r.db.QueryRowContext(ctx, `SELECT id, order_no, merchant_id, account_id, gateway, amount, currency,
		status, customer_email, customer_ip, customer_country, pay_token_hash, gateway_order_id,
		risk_level, risk_score, callback_data, a_site_referer, metadata,
		paid_at, expired_at, canceled_at, created_at, updated_at FROM orders WHERE id=$1`, id).
		Scan(&o.ID, &o.OrderNo, &o.MerchantID, &o.AccountID, &o.Gateway, &o.Amount, &o.Currency,
			&o.Status, &o.CustomerEmail, &o.CustomerIP, &o.CustomerCountry, &o.PayTokenHash, &o.GatewayOrderID,
			&o.RiskLevel, &o.RiskScore, &cb, &o.ASiteReferer, &md,
			&paidAt, &expiredAt, &canceledAt, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(cb, &o.CallbackData)
	json.Unmarshal(md, &o.Metadata)
	if paidAt.Valid {
		o.PaidAt = &paidAt.Time
	}
	if expiredAt.Valid {
		o.ExpiredAt = &expiredAt.Time
	}
	if canceledAt.Valid {
		o.CanceledAt = &canceledAt.Time
	}
	return o, nil
}

func (r *OrderRepo) FindByOrderNo(ctx context.Context, orderNo string) (*model.Order, error) {
	o := &model.Order{}
	var paidAt, expiredAt, canceledAt sql.NullTime
	var cb, md []byte
	err := r.db.QueryRowContext(ctx, `SELECT id, order_no, merchant_id, account_id, gateway, amount, currency,
		status, customer_email, customer_ip, customer_country, pay_token_hash, gateway_order_id,
		risk_level, risk_score, callback_data, a_site_referer, metadata,
		paid_at, expired_at, canceled_at, created_at, updated_at FROM orders WHERE order_no=$1`, orderNo).
		Scan(&o.ID, &o.OrderNo, &o.MerchantID, &o.AccountID, &o.Gateway, &o.Amount, &o.Currency,
			&o.Status, &o.CustomerEmail, &o.CustomerIP, &o.CustomerCountry, &o.PayTokenHash, &o.GatewayOrderID,
			&o.RiskLevel, &o.RiskScore, &cb, &o.ASiteReferer, &md,
			&paidAt, &expiredAt, &canceledAt, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(cb, &o.CallbackData)
	json.Unmarshal(md, &o.Metadata)
	if paidAt.Valid {
		o.PaidAt = &paidAt.Time
	}
	if expiredAt.Valid {
		o.ExpiredAt = &expiredAt.Time
	}
	if canceledAt.Valid {
		o.CanceledAt = &canceledAt.Time
	}
	return o, nil
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id string, status model.OrderStatus, extra map[string]interface{}) error {
	updates := map[string]interface{}{"updated_at": time.Now()}
	updates["status"] = status
	if status == model.StatusPaid {
		now := time.Now()
		updates["paid_at"] = now
	}
	if status == model.StatusCanceled {
		now := time.Now()
		updates["canceled_at"] = now
	}
	if gwID, ok := extra["gateway_order_id"]; ok {
		updates["gateway_order_id"] = gwID
	}
	if gwCID, ok := extra["gateway_customer_id"]; ok {
		updates["gateway_customer_id"] = gwCID
	}
	if cb, ok := extra["callback_data"]; ok {
		data, _ := json.Marshal(cb)
		updates["callback_data"] = data
	}
	setClauses := "SET status=$1, updated_at=$2, paid_at=$3, canceled_at=$4, gateway_order_id=$5, gateway_customer_id=$6, callback_data=$7"
	_, err := r.db.ExecContext(ctx, "UPDATE orders "+setClauses+" WHERE id=$8",
		status, updates["updated_at"], updates["paid_at"], updates["canceled_at"],
		updates["gateway_order_id"], updates["gateway_customer_id"], updates["callback_data"], id)
	return err
}

func (r *OrderRepo) ListByMerchant(ctx context.Context, merchantID string, status string, offset, limit int) ([]*model.Order, error) {
	query := "SELECT id, order_no, merchant_id, account_id, gateway, amount, currency, status, customer_email, customer_country, risk_level, risk_score, created_at, updated_at FROM orders WHERE merchant_id=$1"
	args := []interface{}{merchantID}
	if status != "" {
		query += " AND status=$2"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC LIMIT $3 OFFSET $4"
	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []*model.Order
	for rows.Next() {
		o := &model.Order{}
		rows.Scan(&o.ID, &o.OrderNo, &o.MerchantID, &o.AccountID, &o.Gateway, &o.Amount, &o.Currency, &o.Status, &o.CustomerEmail, &o.CustomerCountry, &o.RiskLevel, &o.RiskScore, &o.CreatedAt, &o.UpdatedAt)
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (r *OrderRepo) CountToday(ctx context.Context, merchantID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders WHERE merchant_id=$1 AND created_at::date = CURRENT_DATE", merchantID).Scan(&count)
	return count, err
}

func (r *OrderRepo) SumToday(ctx context.Context, accountID string) (decimal.Decimal, error) {
	var sum decimal.Decimal
	err := r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(amount), 0) FROM orders WHERE account_id=$1 AND created_at::date = CURRENT_DATE AND status=$2", accountID, "paid").Scan(&sum)
	return sum, err
}
