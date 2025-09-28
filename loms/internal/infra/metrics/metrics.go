package metrics

import (
	"context"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type DBQueryCategory string

const (
	appName = "loms"

	Select DBQueryCategory = "select"
	Create DBQueryCategory = "create"
	Update DBQueryCategory = "update"
	Delete DBQueryCategory = "delete"

	DBQueryStatusOK    = "OK"
	DBQueryStatusError = "error"
)

type Metrics struct {
	requestCounter prometheus.Counter

	grpcRequestTotal             *prometheus.CounterVec
	grpcRequestDurationHistogram *prometheus.HistogramVec

	httpRequestTotal             *prometheus.CounterVec
	httpRequestDurationHistogram *prometheus.HistogramVec

	dbQueryTotal             *prometheus.CounterVec
	dbQueryDurationHistogram *prometheus.HistogramVec

	kafkaProduceTotal             *prometheus.CounterVec
	kafkaProduceDurationHistogram *prometheus.HistogramVec
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
					Help: "The total amout of request",
				},
			),

			grpcRequestTotal: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Name: appName + "_grpc_request_total",
					Help: "The total amount of gRPC requests by path and status",
				},
				[]string{"path", "status"},
			),

			grpcRequestDurationHistogram: promauto.NewHistogramVec(
				prometheus.HistogramOpts{
					Name: appName + "_grpc_request_duration_seconds",
					Help: "Histogram of gRPC requests duration in seconds",
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

			dbQueryTotal: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Name: appName + "_db_query_total",
					Help: "Total number of database queries by category and status",
				},
				[]string{"category", "status"},
			),

			dbQueryDurationHistogram: promauto.NewHistogramVec(
				prometheus.HistogramOpts{
					Name: appName + "_db_query_duration_seconds",
					Help: "Histogram of database query durations by category and status",
					Buckets: []float64{
						0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2,
					},
				},
				[]string{"category", "status"},
			),

			kafkaProduceTotal: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Name: appName + "_kafka_produce_total",
					Help: "Total number of Kafka produce operations by topic and status",
				},
				[]string{"topic", "status"},
			),

			kafkaProduceDurationHistogram: promauto.NewHistogramVec(
				prometheus.HistogramOpts{
					Name: appName + "_kafka_produce_duration_seconds",
					Help: "Histogram of Kafka produce durations by topic and status",
					Buckets: []float64{
						0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5,
					},
				},
				[]string{"topic", "status"},
			),
		}
	})

	return nil
}

func IncRequestCounter() {
	metrics.requestCounter.Inc()
}

func IncGrpcRequestCounter(path, status string) {
	metrics.grpcRequestTotal.WithLabelValues(path, status).Inc()
}

func GrpcRequestDurationHistogram(path, status string, duration float64) {
	metrics.grpcRequestDurationHistogram.WithLabelValues(path, status).Observe(duration)
}

func IncHTTPRequestCounter(path, status string) {
	metrics.httpRequestTotal.WithLabelValues(path, status).Inc()
}

func HTTPRequestDurationHistogram(path, status string, duration float64) {
	metrics.httpRequestDurationHistogram.WithLabelValues(path, status).Observe(duration)
}

func IncDBQueryCounter(category, status string) {
	metrics.dbQueryTotal.WithLabelValues(category, status).Inc()
}

func DBQueryDurationHistogram(category, status string, duration float64) {
	metrics.dbQueryDurationHistogram.WithLabelValues(category, status).Observe(duration)
}

func IncKafkaProduceCounter(topic, status string) {
	metrics.kafkaProduceTotal.WithLabelValues(topic, status).Inc()
}

func KafkaProduceDurationHistogram(topic, status string, duration float64) {
	metrics.kafkaProduceDurationHistogram.WithLabelValues(topic, status).Observe(duration)
}
