package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Order struct {
	ID                 string                 `json:"id" db:"id"`
	OrderNo            string                 `json:"order_no" db:"order_no"`
	MerchantID         string                 `json:"merchant_id" db:"merchant_id"`
	AccountID          string                 `json:"account_id,omitempty" db:"account_id"`
	Gateway            string                 `json:"gateway,omitempty" db:"gateway"`
	Amount             decimal.Decimal        `json:"amount" db:"amount"`
	Currency           string                 `json:"currency" db:"currency"`
	Status             OrderStatus            `json:"status" db:"status"`
	CustomerEmail      string                 `json:"customer_email,omitempty" db:"customer_email"`
	CustomerIP         string                 `json:"customer_ip,omitempty" db:"customer_ip"`
	CustomerCountry    string                 `json:"customer_country,omitempty" db:"customer_country"`
	PayTokenHash       string                 `json:"-" db:"pay_token_hash"`
	GatewayOrderID     string                 `json:"gateway_order_id,omitempty" db:"gateway_order_id"`
	GatewayCustomerID  string                 `json:"gateway_customer_id,omitempty" db:"gateway_customer_id"`
	RiskLevel          string                 `json:"risk_level" db:"risk_level"`
	RiskScore          float64                `json:"risk_score" db:"risk_score"`
	CallbackData       map[string]interface{} `json:"callback_data,omitempty" db:"callback_data"`
	ASiteReferer       string                 `json:"a_site_referer,omitempty" db:"a_site_referer"`
	Metadata           map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	PaidAt             *time.Time             `json:"paid_at,omitempty" db:"paid_at"`
	ExpiredAt          *time.Time             `json:"expired_at,omitempty" db:"expired_at"`
	CanceledAt         *time.Time             `json:"canceled_at,omitempty" db:"canceled_at"`
	CreatedAt          time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at" db:"updated_at"`
}

type OrderStatus string

const (
	StatusPending           OrderStatus = "pending"
	StatusProcessing        OrderStatus = "processing"
	StatusPaid              OrderStatus = "paid"
	StatusFailed            OrderStatus = "failed"
	StatusCanceled          OrderStatus = "canceled"
	StatusExpired           OrderStatus = "expired"
	StatusRefunding         OrderStatus = "refunding"
	StatusRefunded          OrderStatus = "refunded"
	StatusPartiallyRefunded OrderStatus = "partially_refunded"
	StatusDisputed          OrderStatus = "disputed"
	StatusDisputeWon        OrderStatus = "dispute_won"
	StatusDisputeLost       OrderStatus = "dispute_lost"
	StatusCompleted         OrderStatus = "completed"
)

func (s OrderStatus) IsTerminal() bool {
	switch s {
	case StatusCompleted, StatusCanceled, StatusExpired, StatusDisputeWon, StatusDisputeLost, StatusRefunded:
		return true
	}
	return false
}

func (s OrderStatus) IsSuccessful() bool {
	return s == StatusPaid || s == StatusCompleted || s == StatusDisputeWon
}

func (s OrderStatus) CanTransitionTo(next OrderStatus) bool {
	transitions := map[OrderStatus][]OrderStatus{
		StatusPending:    {StatusProcessing, StatusCanceled, StatusExpired},
		StatusProcessing: {StatusPaid, StatusFailed, StatusCanceled},
		StatusPaid:       {StatusRefunding, StatusCompleted, StatusDisputed},
		StatusRefunding:  {StatusRefunded, StatusPartiallyRefunded},
		StatusFailed:     {},
		StatusDisputed:   {StatusDisputeWon, StatusDisputeLost},
		StatusCompleted:  {},
	}
	allowed, ok := transitions[s]
	if !ok {
		return false
	}
	for _, st := range allowed {
		if st == next {
			return true
		}
	}
	return false
}
