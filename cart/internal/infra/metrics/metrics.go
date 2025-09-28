package metrics

import (
	"context"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type StorageQueryCategory string

const (
	appName = "cart"

	Select StorageQueryCategory = "select"
	Create StorageQueryCategory = "create"
	Update StorageQueryCategory = "update"
	Delete StorageQueryCategory = "delete"

	StorageQueryStatusOK = "OK"
	StorageStatusError   = "error"
)

type Metrics struct {
	requestCounter prometheus.Counter

	httpRequestTotal             *prometheus.CounterVec
	httpRequestDurationHistogram *prometheus.HistogramVec

	externalRequestTotal             *prometheus.CounterVec
	externalRequestDurationHistogram *prometheus.HistogramVec

	storageQueryTotal             *prometheus.CounterVec
	storageQueryDurationHistogram *prometheus.HistogramVec

	cartItemCountGauge prometheus.Gauge
}

var (
	metrics  *Metrics
	initOnce sync.Once
)

func Init(_ context.Context) error {
	initOnce.Do(func() {
		metrics = &Metrics{
			requestCounter: promauto.NewCounter(
				prometheus.CounterOpts{
					Name: appName + "_requests_total",
					Help: "The total amount of request",
				},
			),

			httpRequestTotal: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Name: appName + "_http_request_total",
					Help: "The total amount of HTTP requests by path and status",
				},
				[]string{"path", "status"},
			),

			httpRequestDurationHistogram: promauto.NewHistogramVec(
				prometheus.HistogramOpts{
					Name: appName + "_http_request_duration_seconds",
					Help: "Histogram of HTTP requests duration in seconds",
					Buckets: []float64{
						0.1,  // 100ms
						0.2,  // 200ms
						0.25, // 250ms
						0.5,  // 500ms
						1,    // 1s
					},
				},
				[]string{"path", "status"},
			),

			externalRequestTotal: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Name: appName + "_external_request_total",
					Help: "The total amount of external requests by path and status",
				},
				[]string{"path", "status"},
			),

			externalRequestDurationHistogram: promauto.NewHistogramVec(
				prometheus.HistogramOpts{
					Name: appName + "_external_request_duration_seconds",
					Help: "Histogram of external requests duration in seconds",
					Buckets: []float64{
						0.1,  // 100ms
						0.2,  // 200ms
						0.25, // 250ms
						0.5,  // 500ms
						1,    // 1s
					},
				},
				[]string{"path", "status"},
			),

			storageQueryTotal: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Name: appName + "_storage_query_total",
					Help: "Total number of storage queries by category and status",
				},
				[]string{"category", "status"},
			),

			storageQueryDurationHistogram: promauto.NewHistogramVec(
				prometheus.HistogramOpts{
					Name: appName + "_storage_query_duration_seconds",
					Help: "Histogram of storage query durations by category and status",
					Buckets: []float64{
						0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2,
					},
				},
				[]string{"category", "status"},
			),

			cartItemCountGauge: promauto.NewGauge(
				prometheus.GaugeOpts{
					Name: appName + "_storage_items_total",
					Help: "Current number of items in all carts",
				},
			),
		}
	})

	return nil
}

func IncRequestCounter() {
	metrics.requestCounter.Inc()
}

func IncHTTPRequestCounter(path, status string) {
	metrics.httpRequestTotal.WithLabelValues(path, status).Inc()
}

func HTTPRequestDurationHistogram(path, status string, duration float64) {
	metrics.httpRequestDurationHistogram.WithLabelValues(path, status).Observe(duration)
}

func IncExternalRequestCounter(path, status string) {
	metrics.externalRequestTotal.WithLabelValues(path, status).Inc()
}

func ExternalRequestDurationHistogram(path, status string, duration float64) {
	metrics.externalRequestDurationHistogram.WithLabelValues(path, status).Observe(duration)
}

func IncStorageQueryCounter(category, status string) {
	metrics.storageQueryTotal.WithLabelValues(category, status).Inc()
}

func StorageQueryDurationHistogram(category, status string, duration float64) {
	metrics.storageQueryDurationHistogram.WithLabelValues(category, status).Observe(duration)
}

func SetCartItemCount(count uint32) {
	metrics.cartItemCountGauge.Set(float64(count))
}
