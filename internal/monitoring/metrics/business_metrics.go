package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	OrderCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ab_order_created_total",
		Help: "Total number of orders created",
	}, []string{"merchant", "gateway", "currency"})

	OrderPaid = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ab_order_paid_total",
		Help: "Total number of orders paid",
	}, []string{"merchant", "gateway", "currency"})

	OrderFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ab_order_failed_total",
		Help: "Total number of orders failed",
	}, []string{"merchant", "gateway", "reason"})

	OrderAmount = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ab_order_amount_usd",
		Help:    "Order amounts in USD",
		Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 5000},
	}, []string{"merchant", "gateway"})

	AccountHealthGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ab_account_health",
		Help: "Account health status (1=online, 0=offline)",
	}, []string{"account_id", "gateway"})

	AccountDailyAmount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ab_account_daily_amount",
		Help: "Account daily transaction amount",
	}, []string{"account_id"})

	AccountSuccessRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ab_account_success_rate",
		Help: "Account success rate (0-1)",
	}, []string{"account_id"})

	ProxyHealthGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ab_proxy_health",
		Help: "Proxy health status (1=online, 0=offline)",
	}, []string{"proxy_id", "type"})

	ProxyLatency = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ab_proxy_latency_ms",
		Help: "Proxy latency in milliseconds",
	}, []string{"proxy_id"})

	WebhookReceived = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ab_webhook_received_total",
		Help: "Total webhook events received",
	}, []string{"gateway", "event_type"})

	RiskActions = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ab_risk_actions_total",
		Help: "Total risk actions taken",
	}, []string{"action", "risk_level"})

	SchedulerAllocations = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ab_scheduler_allocation_duration_ms",
		Help:    "Scheduler allocation duration in ms",
		Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000},
	}, []string{"strategy", "gateway"})
)
