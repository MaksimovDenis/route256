package api

import (
	"net/http"
	"net/http/httptest"
	"route256/cart/internal/domain"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAndValidateDeleteCartRequest(t *testing.T) {
	t.Parallel()

	s := newMockServer()

	tests := []struct {
		name           string
		userID         string
		expectedErr    error
		expectedUserID uint64
	}{
		{
			name:           "success: api.parseAndValidateDeleteCartRequest",
			userID:         "123",
			expectedErr:    nil,
			expectedUserID: 123,
		},
		{
			name:        "fail: api.parseAndValidateDeleteCartRequest ErrIncorrectUserID",
			userID:      "abc",
			expectedErr: domain.ErrIncorrectUserID,
		},
		{
			name:        "fail: api.parseAndValidateDeleteCartRequest ErrIncorrectUserID",
			userID:      "",
			expectedErr: domain.ErrIncorrectUserID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodDelete, "/user/"+tt.userID+"/cart", nil)
			req.SetPathValue("user_id", tt.userID)

			result, err := s.parseAndValidateDeleteCartRequest(req)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tt.expectedUserID, result.UserID)
			}
		})
	}
}
