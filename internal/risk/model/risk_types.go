package model

import "time"

type RiskLevel string

const (
	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"
	RiskBlock  RiskLevel = "block"
)

type RiskAction string

const (
	ActionAllow       RiskAction = "allow"
	ActionBlock       RiskAction = "block"
	ActionFlag        RiskAction = "flag"
	ActionThrottle    RiskAction = "throttle"
	ActionSwitchAccount RiskAction = "switch_account"
	ActionNotify      RiskAction = "notify"
)

// RiskRule defines a configurable risk detection rule.
type RiskRule struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	Priority    int       `json:"priority"`
	Conditions  []Condition `json:"conditions"`
	Action      RiskAction  `json:"action"`
	RiskLevel   RiskLevel   `json:"risk_level"`
	RiskScore   float64     `json:"risk_score"`
	ThrottleTTL int         `json:"throttle_ttl,omitempty"` // minutes
	NotifyTo    []string    `json:"notify_to,omitempty"`
}

// Condition is a single rule condition.
type Condition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"` // eq, ne, gt, lt, gte, lte, in, contains, regex
	Value    string `json:"value"`
}

// RiskEvent records a risk detection event.
type RiskEvent struct {
	ID         string    `json:"id"`
	MerchantID string    `json:"merchant_id,omitempty"`
	OrderID    string    `json:"order_id,omitempty"`
	AccountID  string    `json:"account_id,omitempty"`
	RuleName   string    `json:"rule_name"`
	RiskLevel  RiskLevel `json:"risk_level"`
	RiskScore  float64   `json:"risk_score"`
	Action     RiskAction `json:"action"`
	Reason     string    `json:"reason"`
	Context    string    `json:"context"` // JSON context data
	CreatedAt  time.Time `json:"created_at"`
}

// RiskProfile defines risk thresholds per merchant.
type RiskProfile struct {
	MaxOrderAmount      float64  `json:"max_order_amount"`
	DailyOrderLimit     int      `json:"daily_order_limit"`
	Require3DS          bool     `json:"require_3ds"`
	CountryBlacklist    []string `json:"country_blacklist"`
	EmailDomainBlacklist []string `json:"email_domain_blacklist"`
	IPBlacklist         []string `json:"ip_blacklist"`
	CardBinBlacklist    []string `json:"card_bin_blacklist"`
	VelocityWindowSec   int      `json:"velocity_window_sec"`
	MaxVelocityCount    int      `json:"max_velocity_count"`
}
