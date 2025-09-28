package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"route256/cart/internal/api/http/handler/utils"
	"route256/cart/internal/domain"

	"github.com/go-playground/validator"
	"github.com/opentracing/opentracing-go"
)

type addItemRequest struct {
	Count uint32 `json:"count" validate:"gt=0,required"`
}

func (s *Server) AddItemHandler(w http.ResponseWriter, r *http.Request) {
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "api.AddItemHandler")
	defer span.Finish()

	req, err := s.parseAndValidateAddItemRequest(r)
	if err != nil {
		makeErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	item := domain.Item{
		Sku:   req.SkuID,
		Count: req.Count,
	}

	err = s.cartService.AddItem(ctx, req.UserID, item)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) ||
			errors.Is(err, domain.ErrNotEnoughStocks) {
			makeErrorResponse(w, err, http.StatusPreconditionFailed)

			return
		}

		makeErrorResponse(w, err, http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}

type addItemParsedRequest struct {
	UserID uint64
	SkuID  domain.Sku
	Count  uint32
}

func (s *Server) parseAndValidateAddItemRequest(r *http.Request) (addItemParsedRequest, error) {
	var req addItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return addItemParsedRequest{}, fmt.Errorf("failed to decode JSON request body: %w", err)
	}

	if err := s.validator.Struct(req); err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			return addItemParsedRequest{}, fmt.Errorf("validation error: %s", formatValidationErrors(valErrs))
		}
		return addItemParsedRequest{}, err
	}

	userIDStr := r.PathValue("user_id")
	userID, err := utils.ConvStrToUint64(userIDStr, domain.ErrIncorrectUserID)
	if err != nil {
		return addItemParsedRequest{}, err
	}

	skuIDStr := r.PathValue("sku_id")
	skuID, err := utils.ConvStrToUint64(skuIDStr, domain.ErrIncorrectSku)
	if err != nil {
		return addItemParsedRequest{}, err
	}

	return addItemParsedRequest{
		UserID: userID,
		SkuID:  domain.Sku(skuID),
		Count:  req.Count,
	}, nil
}
