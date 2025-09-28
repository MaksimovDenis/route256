package api

import (
	"net/http"
	"route256/cart/internal/api/http/handler/utils"
	"route256/cart/internal/domain"

	"github.com/opentracing/opentracing-go"
)

func (s *Server) DeleteCartByUserID(w http.ResponseWriter, r *http.Request) {
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "api.DeleteCartByUserID")
	defer span.Finish()

	req, err := s.parseAndValidateDeleteCartRequest(r)
	if err != nil {
		makeErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	err = s.cartService.DeleteItemsByUserID(ctx, req.UserID)
	if err != nil {
		makeErrorResponse(w, err, http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type deleteCartParsedRequest struct {
	UserID uint64
}

func (s *Server) parseAndValidateDeleteCartRequest(r *http.Request) (deleteCartParsedRequest, error) {
	userIDStr := r.PathValue("user_id")
	userID, err := utils.ConvStrToUint64(userIDStr, domain.ErrIncorrectUserID)
	if err != nil {
		return deleteCartParsedRequest{}, err
	}

	return deleteCartParsedRequest{
		UserID: userID,
	}, nil
}
