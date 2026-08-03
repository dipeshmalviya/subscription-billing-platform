package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ChargesTotal counts every charge attempt, labeled by outcome
	// (succeeded/failed), so you can graph success rate over time.
	ChargesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "payment_charges_total",
		Help: "Total number of charge attempts, labeled by outcome",
	}, []string{"status"})

	// RetryAttemptsTotal counts how many times the background retry worker
	// re-attempted a previously failed payment.
	RetryAttemptsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "payment_retry_attempts_total",
		Help: "Total number of retry-worker charge attempts",
	})

	// RefundsTotal counts refund requests, labeled by outcome.
	RefundsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "payment_refunds_total",
		Help: "Total number of refund attempts, labeled by outcome",
	}, []string{"status"})

	// ChargeDuration tracks how long the mock gateway call takes — in a real
	// integration this is what would catch a slow/degraded payment provider.
	ChargeDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "payment_charge_duration_seconds",
		Help:    "Duration of gateway charge calls",
		Buckets: prometheus.DefBuckets,
	})

	// FailedPaymentsPendingRetry is a gauge showing current backlog size —
	// useful for alerting if the retry worker falls behind.
	FailedPaymentsPendingRetry = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "payment_failed_pending_retry",
		Help: "Number of failed payments currently eligible for retry",
	})
)