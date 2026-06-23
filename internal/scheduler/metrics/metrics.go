package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	AllocationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ab_allocation_total",
		Help: "Total number of payment account allocations",
	}, []string{"gateway", "strategy", "result"})

	AllocationLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ab_allocation_latency_seconds",
		Help:    "Allocation latency in seconds",
		Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
	}, []string{"gateway"})

	AccountStatusGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ab_account_status",
		Help: "Number of accounts by gateway and status",
	}, []string{"gateway", "status"})

	OrdersTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ab_orders_total",
		Help: "Total order count by status and gateway",
	}, []string{"status", "gateway", "merchant"})
)
