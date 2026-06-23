package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Merchant struct {
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	Email         string           `json:"email"`
	Status        string           `json:"status"` // active | suspended | banned
	RiskProfile   *RiskProfile     `json:"risk_profile,omitempty"`
	FeeConfig     *FeeConfig       `json:"fee_config,omitempty"`
	Settlement    *SettlementConfig `json:"settlement,omitempty"`
	RoutingMode   string           `json:"routing_mode"` // weighted_rr | sequential | random | least_utilized
	AccountGroup  string           `json:"account_group"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

type FeeConfig struct {
	PayPalFee *GatewayFee `json:"paypal_fee"`
	StripeFee *GatewayFee `json:"stripe_fee"`
}

type GatewayFee struct {
	FixedFee decimal.Decimal `json:"fixed_fee"` // e.g. $0.30
	RateFee  decimal.Decimal `json:"rate_fee"`  // e.g. 0.045 (4.5%)
	MinFee   decimal.Decimal `json:"min_fee"`   // minimum fee
}

type SettlementConfig struct {
	Method       string          `json:"method"`        // USDT | wire | wise
	MinAmount    decimal.Decimal `json:"min_amount"`     // minimum settlement amount
	Cycle        string          `json:"cycle"`          // daily | weekly | threshold
	Address      string          `json:"address"`        // payout address
	Currency     string          `json:"currency"`
}

type RiskProfile struct {
	MaxOrderAmount      float64  `json:"max_order_amount"`
	DailyOrderLimit     int      `json:"daily_order_limit"`
	CountryBlacklist    []string `json:"country_blacklist"`
	Require3DS          bool     `json:"require_3ds"`
}

type APIKey struct {
	KeyID       string     `json:"key_id"`
	KeyHash     string     `json:"-"` // SHA256(secret), never serialized
	KeyPrefix   string     `json:"key_prefix"` // first 8 chars
	Permissions []string   `json:"permissions"` // read | write | admin
	IPWhitelist []string   `json:"ip_whitelist"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
}

type ASite struct {
	Domain     string `json:"domain"`
	MerchantID string `json:"merchant_id"`
	Active     bool   `json:"active"`
}
