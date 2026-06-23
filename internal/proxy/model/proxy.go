package model

import "time"

type Proxy struct {
	ID                string    `json:"id"`
	Type              string    `json:"type"`     // residential | datacenter | mobile | fixed_isp
	Protocol          string    `json:"protocol"` // http | socks5
	Host              string    `json:"host"`
	Port              int       `json:"port"`
	Username          string    `json:"username,omitempty"`
	EncryptedPassword []byte    `json:"-"`
	Country           string    `json:"country"`
	City              string    `json:"city"`
	ISP               string    `json:"isp"`

	BoundAccountID string `json:"bound_account_id"`
	IsDedicated    bool   `json:"is_dedicated"`

	Status      string  `json:"status"` // online | offline | testing | banned
	Latency     int     `json:"latency"`    // milliseconds
	SuccessRate float64 `json:"success_rate"`
	UsedCount   int64   `json:"used_count"`

	BandwidthLimit int64 `json:"bandwidth_limit"`
	BandwidthUsed  int64 `json:"bandwidth_used"`

	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProxyHealth struct {
	Latency          int  `json:"latency"`
	Blacklisted      bool `json:"blacklisted"`
	GeoMatch         bool `json:"geo_match"`
	PayPalAccessible bool `json:"paypal_accessible"`
	StripeAccessible bool `json:"stripe_accessible"`
	IsHealthy        bool `json:"is_healthy"`
}
