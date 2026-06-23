package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Settlement struct {
	ID           string          `json:"id"`
	MerchantID   string          `json:"merchant_id"`
	CycleStart   time.Time       `json:"cycle_start"`
	CycleEnd     time.Time       `json:"cycle_end"`
	TotalOrders  int             `json:"total_orders"`
	TotalAmount  decimal.Decimal `json:"total_amount"`
	TotalFee     decimal.Decimal `json:"total_fee"`
	NetAmount    decimal.Decimal `json:"net_amount"`
	PayoutMethod string          `json:"payout_method"` // USDT | wire | wise
	PayoutAddress string          `json:"payout_address"`
	PayoutTxID   string          `json:"payout_tx_id,omitempty"`
	Status       string          `json:"status"` // pending | processing | completed | failed
	SettledAt    *time.Time      `json:"settled_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type SettlementConfig struct {
	Method    string          `json:"method"`
	MinAmount decimal.Decimal `json:"min_amount"`
	Cycle     string          `json:"cycle"` // daily | weekly | threshold
	Address   string          `json:"address"`
}

type PayoutRequest struct {
	SettlementID string          `json:"settlement_id"`
	Method       string          `json:"method"`
	Address      string          `json:"address"`
	Amount       decimal.Decimal `json:"amount"`
	Currency     string          `json:"currency"`
	Reference    string          `json:"reference"`
}
