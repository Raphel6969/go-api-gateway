package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	RequestsTotal    *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	InFlightRequests prometheus.Gauge
	CacheHits        prometheus.Counter
	CacheMisses      prometheus.Counter
}

func New() *Metrics {
	return &Metrics{
		RequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_requests_total",
				Help: "Total number of HTTP requests processed by the gateway",
			},
			[]string{"method", "path", "status"},
		),
		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_requests_duration_seconds",
				Help:    "Histogram of latencies for HTTP requests processed by the gateway",
				Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
			},
			[]string{"method", "path"},
		),
		InFlightRequests: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "gateway_in_flight_requests",
				Help: "Current number of in-flight requests handled by the gateway",
			},
		),
		CacheHits: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "gateway_cache_hits_total",
				Help: "Total number of cache hits in the gateway cache engine",
			},
		),
		CacheMisses: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "gateway_cache_misses_total",
				Help: "Total number of cache misses in the gateway cache engine",
			},
		),
	}
}
