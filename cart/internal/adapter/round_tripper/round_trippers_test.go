package roundtripper_test

import (
	"context"
	"fmt"
	"net/http"
	roundtripper "route256/cart/internal/adapter/round_tripper"
	"route256/cart/internal/infra/metrics"

	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRoundTripper struct {
	response  []*http.Response
	errs      []error
	callCount int
}

func (m *mockRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	count := m.callCount
	m.callCount++

	if count < len(m.errs) {
		return nil, m.errs[count]
	}

	if count < len(m.response) {
		return m.response[count], nil
	}

	return &http.Response{StatusCode: 200}, nil
}

func TestRetryRoundTripper(t *testing.T) {
	t.Parallel()
	err := metrics.Init(context.Background())
	require.NoError(t, err)

	tests := []struct {
		name          string
		mockResp      []*http.Response
		mockErrs      []error
		maxRetries    int
		expectedCode  int
		expectedError error
	}{
		{
			name: "success: roundtripper.RoundTrip - default case",
			mockResp: []*http.Response{
				{StatusCode: 200},
			},
			maxRetries:    3,
			expectedCode:  200,
			expectedError: nil,
		},
		{
			name: "success: roundtripper.RoundTrip - retry on 429",
			mockResp: []*http.Response{
				{StatusCode: 429},
				{StatusCode: 200},
			},
			maxRetries:    3,
			expectedCode:  200,
			expectedError: nil,
		},
		{
			name: "success: roundtripper.RoundTrip - retry on 420",
			mockResp: []*http.Response{
				{StatusCode: 420},
				{StatusCode: 200},
			},
			maxRetries:    3,
			expectedCode:  200,
			expectedError: nil,
		},
		{
			name: "fail: roundtripper.RoundTrip - max retries reached",
			mockResp: []*http.Response{
				{StatusCode: 420},
			},
			maxRetries:    1,
			expectedCode:  420,
			expectedError: fmt.Errorf("max retries reached: %v, last response code: %v", 1, 420),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rt := roundtripper.New(
				&mockRoundTripper{
					response: tt.mockResp,
					errs:     tt.mockErrs,
				},
				tt.maxRetries,
				10*time.Millisecond,
			)

			req, err := http.NewRequest(
				http.MethodGet,
				"http://example.ru",
				nil,
			)
			require.NoError(t, err)

			resp, err := rt.RoundTrip(req)

			if tt.expectedError != nil {
				require.ErrorContains(t, err, tt.expectedError.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedCode, resp.StatusCode)
			}
		})
	}
}
