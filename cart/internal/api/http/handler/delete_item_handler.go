package api

import (
	"net/http"
	"route256/cart/internal/api/http/handler/utils"
	"route256/cart/internal/domain"

	"github.com/opentracing/opentracing-go"
)

func (s *Server) DeleteItemHandler(w http.ResponseWriter, r *http.Request) {
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "api.DeleteItemHandler")
	defer span.Finish()

	req, err := s.parseAndValidateDeleteItemRequest(r)
	if err != nil {
		makeErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	err = s.cartService.DeleteItem(ctx, req.UserID, req.SkuID)
	if err != nil {
		makeErrorResponse(w, err, http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type deleteItemParsedRequest struct {
	UserID uint64
	SkuID  domain.Sku
}

func (s *Server) parseAndValidateDeleteItemRequest(r *http.Request) (deleteItemParsedRequest, error) {
	userIDStr := r.PathValue("user_id")
	userID, err := utils.ConvStrToUint64(userIDStr, domain.ErrIncorrectUserID)
	if err != nil {
		return deleteItemParsedRequest{}, err
	}

	skuIDStr := r.PathValue("sku_id")
	skuID, err := utils.ConvStrToUint64(skuIDStr, domain.ErrIncorrectSku)
	if err != nil {
		return deleteItemParsedRequest{}, err
	}

	return deleteItemParsedRequest{
		UserID: userID,
		SkuID:  domain.Sku(skuID),
	}, nil
}
