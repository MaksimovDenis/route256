package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"route256/cart/internal/domain"
	"testing"

	"github.com/go-playground/validator"
	"github.com/stretchr/testify/require"
)

type mockServer struct {
	*Server
}

func newMockServer() *mockServer {
	validate := validator.New()
	return &mockServer{
		Server: &Server{
			validator: validate,
		},
	}
}

func TestParseAndValidateAddItemRequest(t *testing.T) {
	t.Parallel()

	var (
		s = newMockServer()
	)

	tests := []struct {
		name           string
		body           string
		userID         string
		skuID          string
		expectedErr    error
		expectedCount  uint32
		expectedUserID uint64
		expectedSkuID  domain.Sku
	}{
		{
			name:           "success: api.parseAndValidateAddItemRequest",
			body:           `{"count": 5}`,
			userID:         "123",
			skuID:          "456",
			expectedErr:    nil,
			expectedCount:  5,
			expectedUserID: 123,
			expectedSkuID:  456,
		},
		{
			name:        "fail: api.parseAndValidateAddItemRequest Invalid json",
			body:        `{"count": "bad"}`,
			userID:      "123",
			skuID:       "456",
			expectedErr: fmt.Errorf("failed to decode JSON request body"),
		},
		{
			name:        "fail: api.parseAndValidateAddItemRequest ErrIncorrectUserID",
			body:        `{"count": 1}`,
			userID:      "abc",
			skuID:       "456",
			expectedErr: domain.ErrIncorrectUserID,
		},
		{
			name:        "fail: api.parseAndValidateAddItemRequest ErrIncorrectSku",
			body:        `{"count": 1}`,
			userID:      "123",
			skuID:       "zzz",
			expectedErr: domain.ErrIncorrectSku,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/user/%s/cart/%s", tt.userID, tt.skuID), bytes.NewBufferString(tt.body))
			req.SetPathValue("user_id", tt.userID)
			req.SetPathValue("sku_id", tt.skuID)

			result, err := s.parseAndValidateAddItemRequest(req)

			if tt.expectedErr != nil {
				require.ErrorContains(t, err, tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tt.expectedCount, result.Count)
				require.Equal(t, tt.expectedUserID, result.UserID)
				require.Equal(t, tt.expectedSkuID, result.SkuID)
			}
		})
	}
}
