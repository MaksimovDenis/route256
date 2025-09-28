package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"route256/cart/internal/api/http/handler/utils"
	"route256/cart/internal/domain"

	"github.com/opentracing/opentracing-go"
)

type GetItemsByUserIDResItem struct {
	Sku   uint64 `json:"sku"`
	Name  string `json:"name"`
	Count uint32 `json:"count"`
	Price uint32 `json:"price"`
}

type GetItemsByUserIDRes struct {
	Items      []GetItemsByUserIDResItem `json:"items"`
	TotalPrice uint32                    `json:"total_price"`
}

func (s *Server) GetItemsByUserID(w http.ResponseWriter, r *http.Request) {
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "api.GetItemsByUserID")
	defer span.Finish()

	req, err := s.parseAndValidateGetItemRequest(r)
	if err != nil {
		makeErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	cart, err := s.cartService.GetItemsByUserID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrEmptyCart) {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		makeErrorResponse(w, err, http.StatusInternalServerError)

		return
	}

	items := make([]GetItemsByUserIDResItem, len(cart.Items))

	for idx, item := range cart.Items {
		items[idx] = GetItemsByUserIDResItem{
			Sku:   uint64(item.Item.Sku),
			Name:  item.Name,
			Count: item.Count,
			Price: item.Price,
		}
	}

	response := &GetItemsByUserIDRes{
		Items:      items,
		TotalPrice: cart.TotalPrice,
	}

	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		makeErrorResponse(w, err, http.StatusInternalServerError)

		return
	}
}

type getItemParsedRequest struct {
	UserID uint64
}

func (s *Server) parseAndValidateGetItemRequest(r *http.Request) (getItemParsedRequest, error) {
	userIDStr := r.PathValue("user_id")
	userID, err := utils.ConvStrToUint64(userIDStr, domain.ErrIncorrectUserID)
	if err != nil {
		return getItemParsedRequest{}, err
	}

	return getItemParsedRequest{
		UserID: userID,
	}, nil
}
