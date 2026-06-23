package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type AllocateRequest struct {
	OrderID    string
	Amount     decimal.Decimal
	Currency   string
	MerchantID string
	Gateway    string
	Strategy   string
	Country    string
	Metadata   map[string]string
}

type AllocateResponse struct {
	PayToken   string
	GatewayURL string
	AccountRef string
	ExpiresAt  time.Time
	Gateway    string
}

type ReleaseRequest struct {
	OrderID    string
	MerchantID string
}

type ReleaseResponse struct {
	Success bool
}

type ConfirmRequest struct {
	OrderID    string
	MerchantID string
}

type ConfirmResponse struct {
	Success bool
}

type SystemHealth struct {
	Healthy           bool
	OnlineAccounts    int
	TotalOrdersToday  int
	AvgLatencyMs      float64
}
