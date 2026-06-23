package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// Account represents a payment gateway account (PayPal/Stripe)
type Account struct {
	ID            string      `json:"id" db:"id"`
	Gateway       string      `json:"gateway" db:"gateway"` // paypal | stripe
	Alias         string      `json:"alias" db:"alias"`
	Status        string      `json:"status" db:"status"` // online|cooling|offline|draining|warming
	EncryptedCred []byte      `json:"-" db:"encrypted_cred"`
	Credential    *Credential `json:"-"` // decrypted in memory only
	BSiteID       string      `json:"b_site_id" db:"b_site_id"`
	MerchantID    string      `json:"merchant_id" db:"merchant_id"`

	LimitConfig LimitConfig `json:"limit_config" db:"-"`

	Weight    int      `json:"weight" db:"weight"`
	Priority  int      `json:"priority" db:"priority"`
	Tags      []string `json:"tags" db:"tags"`

	SupportedCurrencies []string `json:"supported_currencies" db:"-"`
	SupportedCountries  []string `json:"supported_countries" db:"-"`

	Runtime *AccountRuntime `json:"-"` // runtime state from Redis

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Credential holds decrypted payment gateway API credentials
type Credential struct {
	PayPalClientID     string `json:"paypal_client_id,omitempty"`
	PayPalSecret       string `json:"paypal_secret,omitempty"`
	PayPalMerchantID   string `json:"paypal_merchant_id,omitempty"`
	PayPalWebhookID    string `json:"paypal_webhook_id,omitempty"`

	StripePublishableKey string `json:"stripe_publishable_key,omitempty"`
	StripeSecretKey      string `json:"stripe_secret_key,omitempty"`
	StripeAccountID      string `json:"stripe_account_id,omitempty"`
	StripeWebhookSecret  string `json:"stripe_webhook_secret,omitempty"`
}

// LimitConfig defines the transaction limits for an account
type LimitConfig struct {
	SingleMin   decimal.Decimal `json:"single_min"`
	SingleMax   decimal.Decimal `json:"single_max"`
	DailyMax    decimal.Decimal `json:"daily_max"`
	MonthlyMax  decimal.Decimal `json:"monthly_max"`
	LifetimeMax decimal.Decimal `json:"lifetime_max"`
	DailyCount  int             `json:"daily_count"`
}

// AccountRuntime holds the runtime state of an account (stored in Redis)
type AccountRuntime struct {
	AccountID        string          `json:"account_id"`
	Status           string          `json:"status"`
	TodayAmount      decimal.Decimal `json:"today_amount"`
	TodayCount       int             `json:"today_count"`
	MonthAmount      decimal.Decimal `json:"month_amount"`
	LifetimeAmount   decimal.Decimal `json:"lifetime_amount"`
	LastUsedAt       time.Time       `json:"last_used_at"`
	ConsecutiveFails int             `json:"consecutive_fails"`
	SuccessRate      float64         `json:"success_rate"`
	AvgLatency       int             `json:"avg_latency_ms"`
	CoolingUntil     *time.Time      `json:"cooling_until,omitempty"`
}

// BSite represents a B-site (普货网站)
type BSite struct {
	ID       string `json:"id"`
	Domain   string `json:"domain"`
	Name     string `json:"name"`
	HostingIP string `json:"hosting_ip"`
	Status   string `json:"status"`
}
