package roundtripper

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"route256/cart/internal/infra/metrics"
	"time"
)

type RetryRoundTripper struct {
	rt           http.RoundTripper
	maxRetries   int
	retryTimeout time.Duration
}

func New(
	rt http.RoundTripper,
	maxRetries int,
	retryTimeout time.Duration,
) *RetryRoundTripper {
	return &RetryRoundTripper{
		rt:           rt,
		maxRetries:   maxRetries,
		retryTimeout: retryTimeout,
	}
}

func (r *RetryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("roundtripper.RoundTrip read request body: %w", err)
		}

		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	start := time.Now()

	for attempts := 0; attempts < r.maxRetries; attempts++ {
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		resp, err = r.rt.RoundTrip(req)
		if err != nil {
			metrics.IncExternalRequestCounter(req.URL.Path, "error")
			metrics.ExternalRequestDurationHistogram(req.URL.Path, "error", time.Since(start).Seconds())
			return nil, err
		}

		statusCodeStr := fmt.Sprintf("%d", resp.StatusCode)
		metrics.IncExternalRequestCounter(req.URL.Path, statusCodeStr)
		metrics.ExternalRequestDurationHistogram(req.URL.Path, statusCodeStr, time.Since(start).Seconds())

		if resp.StatusCode == 420 || resp.StatusCode == 429 {
			time.Sleep(r.retryTimeout)
			continue
		}

		return resp, nil
	}

	statusCodeStr := fmt.Sprintf("%d", resp.StatusCode)
	metrics.IncExternalRequestCounter(req.URL.Path, statusCodeStr)
	metrics.ExternalRequestDurationHistogram(req.URL.Path, statusCodeStr, time.Since(start).Seconds())

	return resp, fmt.Errorf("max retries reached: %d, last response code: %d", r.maxRetries, resp.StatusCode)
}
