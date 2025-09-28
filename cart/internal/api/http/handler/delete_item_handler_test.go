package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"route256/cart/internal/domain"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAndValidateDeleteItemRequest(t *testing.T) {
	t.Parallel()

	s := newMockServer()

	tests := []struct {
		name           string
		userID         string
		skuID          string
		expectedErr    error
		expectedUserID uint64
		expectedSkuID  domain.Sku
	}{
		{
			name:           "success: api.parseAndValidateDeleteItemRequest",
			userID:         "123",
			skuID:          "456",
			expectedErr:    nil,
			expectedUserID: 123,
			expectedSkuID:  456,
		},
		{
			name:        "fail: api.parseAndValidateDeleteItemRequest ErrIncorrectUserID",
			userID:      "abc",
			skuID:       "456",
			expectedErr: domain.ErrIncorrectUserID,
		},
		{
			name:        "fail: api.parseAndValidateDeleteItemRequest ErrIncorrectSku",
			userID:      "123",
			skuID:       "xyz",
			expectedErr: domain.ErrIncorrectSku,
		},
		{
			name:        "fail: api.parseAndValidateDeleteItemRequest ErrIncorrectUserID",
			userID:      "",
			skuID:       "789",
			expectedErr: domain.ErrIncorrectUserID,
		},
		{
			name:        "fail: api.parseAndValidateDeleteItemRequest ErrIncorrectSku",
			userID:      "123",
			skuID:       "",
			expectedErr: domain.ErrIncorrectSku,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/user/%s/cart/%s", tt.userID, tt.skuID), nil)
			req.SetPathValue("user_id", tt.userID)
			req.SetPathValue("sku_id", tt.skuID)

			result, err := s.parseAndValidateDeleteItemRequest(req)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tt.expectedUserID, result.UserID)
				require.Equal(t, tt.expectedSkuID, result.SkuID)
			}
		})
	}
}
