package middleware

import (
	"net/http"
	"route256/loms/internal/infra/metrics"
	"time"

	"github.com/opentracing/opentracing-go"
)

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusResponseWriter) Write(data []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}

	return w.ResponseWriter.Write(data)
}

func HTTPMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.IncRequestCounter()

		start := time.Now()

		wrapped := &statusResponseWriter{ResponseWriter: w}

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
