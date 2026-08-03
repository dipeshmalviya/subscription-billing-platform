package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var NotificationsSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "notifications_sent_total",
	Help: "Total notifications sent, labeled by event type",
}, []string{"event_type"})