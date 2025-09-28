package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"route256/cart/internal/domain"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAndValidateGetItemRequest(t *testing.T) {
	t.Parallel()

	s := newMockServer()

	tests := []struct {
		name           string
		userID         string
		expectedErr    error
		expectedUserID uint64
	}{
		{
			name:           "success: api.parseAndValidateDeleteItemRequest",
			userID:         "123",
			expectedUserID: 123,
		},
		{
			name:        "fail: api.parseAndValidateDeleteItemRequest ErrIncorrectUserID",
			userID:      "abc",
			expectedErr: domain.ErrIncorrectUserID,
		},
		{
			name:        "fail: api.parseAndValidateDeleteItemRequest ErrIncorrectUserID",
			userID:      "",
			expectedErr: domain.ErrIncorrectUserID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/user/%s/cart", tt.userID), nil)
			req.SetPathValue("user_id", tt.userID)

			result, err := s.parseAndValidateGetItemRequest(req)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tt.expectedUserID, result.UserID)
			}
		})
	}
}
