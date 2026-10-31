package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "marketdata",
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "Total HTTP requests",
	}, []string{"method", "path", "status"})

	HTTPDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "marketdata",
		Subsystem: "http",
		Name:      "request_duration_seconds",
		Help:      "HTTP request duration",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "path"})

	ProviderRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "marketdata",
		Subsystem: "provider",
		Name:      "requests_total",
		Help:      "Provider API requests",
	}, []string{"provider", "method", "status"})

	ProviderDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "marketdata",
		Subsystem: "provider",
		Name:      "request_duration_seconds",
		Help:      "Provider API request duration",
		Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"provider", "method"})
)
