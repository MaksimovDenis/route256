package roundtripper

import (
	"fmt"
	"net/http"

	"golang.org/x/time/rate"
)

type RateLimitRoundTripper struct {
	roundTripperWrap http.RoundTripper
	ratelimiter      *rate.Limiter
}

func NewRateLimitRoundTripper(rt http.RoundTripper, rl *rate.Limiter) *RateLimitRoundTripper {
	return &RateLimitRoundTripper{
		roundTripperWrap: rt,
		ratelimiter:      rl,
	}
}

func (t *RateLimitRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := t.ratelimiter.Wait(req.Context()); err != nil {
		return nil, fmt.Errorf("rate limiter wait: %w", err)
	}

	return t.roundTripperWrap.RoundTrip(req)
}
