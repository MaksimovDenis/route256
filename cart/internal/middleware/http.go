package middleware

import (
	"net/http"
	"route256/cart/internal/infra/logger"
	"route256/cart/internal/infra/metrics"
	"time"

	"github.com/opentracing/opentracing-go"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

func NewLoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rr := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

			if r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()

			next.ServeHTTP(rr, r)

			duration := time.Since(start)

			logger.LogRequest(
				r.Method,
				r.URL.String(),
				rr.statusCode,
				duration,
			)
		})
	}
}

func HTTPMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseRecorder{ResponseWriter: w}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		path := r.URL.Path
		code := wrapped.statusCode

		metrics.IncHTTPRequestCounter(path, http.StatusText(code))
		metrics.HTTPRequestDurationHistogram(path, http.StatusText(code), duration)
	})
}

func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spanCtx, _ := opentracing.GlobalTracer().Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header),
		)

		span := opentracing.StartSpan(
			r.URL.Path,
			opentracing.ChildOf(spanCtx),
		)
		defer span.Finish()

		ctx := opentracing.ContextWithSpan(r.Context(), span)
		r = r.WithContext(ctx)

		span.SetTag("http.method", r.Method)
		span.SetTag("http.url", r.URL.String())

		next.ServeHTTP(w, r)
	})
}
