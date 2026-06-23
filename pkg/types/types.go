// Package types provides shared type definitions across all modules
package types

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

// AccountStatus represents the lifecycle status of a payment account
type AccountStatus string

const (
	AccountStatusOnline   AccountStatus = "online"
	AccountStatusCooling  AccountStatus = "cooling"
	AccountStatusOffline  AccountStatus = "offline"
	AccountStatusDraining AccountStatus = "draining"
	AccountStatusWarming  AccountStatus = "warming"
)

// GatewayType identifies the payment gateway provider
type GatewayType string

const (
	GatewayPayPal GatewayType = "paypal"
	GatewayStripe GatewayType = "stripe"
)

// OrderStatus tracks the lifecycle of a payment order
type OrderStatus string

const (
	StatusPending          OrderStatus = "pending"
	StatusProcessing       OrderStatus = "processing"
	StatusPaid             OrderStatus = "paid"
	StatusFailed           OrderStatus = "failed"
	StatusCanceled         OrderStatus = "canceled"
	StatusExpired          OrderStatus = "expired"
	StatusRefunding        OrderStatus = "refunding"
	StatusRefunded         OrderStatus = "refunded"
	StatusPartiallyRefunded OrderStatus = "partially_refunded"
	StatusDisputed         OrderStatus = "disputed"
	StatusDisputeWon       OrderStatus = "dispute_won"
	StatusDisputeLost      OrderStatus = "dispute_lost"
	StatusRetrying         OrderStatus = "retrying"
	StatusCompleted        OrderStatus = "completed"
)

// ProxyType classifies proxy sources
type ProxyType string

const (
	ProxyResidential ProxyType = "residential"
	ProxyDatacenter  ProxyType = "datacenter"
	ProxyMobile      ProxyType = "mobile"
	ProxyFixedISP    ProxyType = "fixed_isp"
)

// ProxyStatus tracks proxy health
type ProxyStatus string

const (
	ProxyStatusOnline  ProxyStatus = "online"
	ProxyStatusOffline ProxyStatus = "offline"
	ProxyStatusTesting ProxyStatus = "testing"
	ProxyStatusBanned  ProxyStatus = "banned"
)

// RiskLevel for risk assessment
type RiskLevel string

const (
	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"
	RiskBlock  RiskLevel = "block"
)

// MerchantStatus for merchant accounts
type MerchantStatus string

const (
	MerchantActive    MerchantStatus = "active"
	MerchantSuspended MerchantStatus = "suspended"
	MerchantBanned    MerchantStatus = "banned"
)

// SettlementStatus for settlement cycles
type SettlementStatus string

const (
	SettlementPending    SettlementStatus = "pending"
	SettlementProcessing SettlementStatus = "processing"
	SettlementCompleted  SettlementStatus = "completed"
	SettlementFailed     SettlementStatus = "failed"
)

// Currency supported
type Currency string

const (
	USD Currency = "USD"
	EUR Currency = "EUR"
	GBP Currency = "GBP"
	CAD Currency = "CAD"
	AUD Currency = "AUD"
	JPY Currency = "JPY"
	HKD Currency = "HKD"
)

// RoutingStrategy determines how accounts are selected
type RoutingStrategy string

const (
	StrategyWeightedRR    RoutingStrategy = "weighted_round_robin"
	StrategySequential    RoutingStrategy = "sequential"
	StrategyRandom        RoutingStrategy = "random"
	StrategyLeastUtilized RoutingStrategy = "least_utilized"
	StrategyPriority      RoutingStrategy = "priority"
)

// ========== Decimal type alias for convenience ==========

// Decimal is a type alias for shopspring/decimal.Decimal
type Decimal = decimal.Decimal

// ========== Common JSON types ==========

// JSONMap is a convenience type for raw JSON data
type JSONMap map[string]interface{}

// Value implements the driver.Valuer interface for database serialization
func (jm JSONMap) Value() ([]byte, error) {
	return json.Marshal(jm)
}

// Scan implements the sql.Scanner interface for database deserialization
func (jm *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*jm = make(JSONMap)
		return nil
	}
	return json.Unmarshal(value.([]byte), jm)
}

// ========== Time utilities ==========

// Now returns the current time in UTC
func Now() time.Time { return time.Now().UTC() }

// NowPtr returns a pointer to the current time in UTC
func NowPtr() *time.Time {
	t := time.Now().UTC()
	return &t
}
