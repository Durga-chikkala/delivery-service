package helpers

import (
	"sync"

	"github.com/Durga-Chikkala/delivery-service/models"
	"github.com/prometheus/client_golang/prometheus"
)

var metricsOnce sync.Once

func NewMetrics() *models.Metrics {
	m := &models.Metrics{
		RequestCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_requests_total",
				Help: "Total number of API requests.",
			},
			[]string{"method", "endpoint"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "api_request_duration_seconds",
				Help:    "Histogram of latencies for API requests.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		ErrorCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_errors_total",
				Help: "Total number of API errors.",
			},
			[]string{"method", "endpoint", "statusCode"},
		),
		CacheHits: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_hits_total",
				Help: "Total number of cache hits.",
			},
			[]string{"cache_name"},
		),
		CacheMisses: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_misses_total",
				Help: "Total number of cache misses.",
			},
			[]string{"cache_name"},
		),
	}

	metricsOnce.Do(func() {
		prometheus.MustRegister(m.RequestCounter)
		prometheus.MustRegister(m.RequestDuration)
		prometheus.MustRegister(m.ErrorCounter)
		prometheus.MustRegister(m.CacheHits)
		prometheus.MustRegister(m.CacheMisses)
	})

	return m
}
