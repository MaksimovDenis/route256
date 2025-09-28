package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"route256/cart/internal/api/http/handler/utils"
	"route256/cart/internal/domain"

	"github.com/opentracing/opentracing-go"
)

type checkoutResponse struct {
	OrderID int64 `json:"order_id"`
}

func (s *Server) CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "api.CheckoutHandler")
	defer span.Finish()

	req, err := s.parseAndValidateCheckoutRequest(r)
	if err != nil {
		makeErrorResponse(w, err, http.StatusBadRequest)

		return
	}

	orderID, err := s.cartService.Checkout(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrEmptyCart) {
			makeErrorResponse(w, err, http.StatusNotFound)

			return
		}
		makeErrorResponse(w, err, http.StatusInternalServerError)

		return
	}

	response := &checkoutResponse{
		OrderID: orderID,
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		makeErrorResponse(w, err, http.StatusInternalServerError)

		return
	}
}

type checkoutParsedRequest struct {
	UserID uint64
}

func (s *Server) parseAndValidateCheckoutRequest(r *http.Request) (checkoutParsedRequest, error) {
	userIDStr := r.PathValue("user_id")
	userID, err := utils.ConvStrToUint64(userIDStr, domain.ErrIncorrectUserID)
	if err != nil {
		return checkoutParsedRequest{}, err
	}

	return checkoutParsedRequest{
		UserID: userID,
	}, nil
}
