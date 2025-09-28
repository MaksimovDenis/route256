package api

import (
	"context"
	"net/http"
	"route256/cart/internal/domain"
	"route256/cart/internal/middleware"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type CartService interface {
	GetItemsByUserID(ctx context.Context, userID uint64) (domain.Cart, error)
	AddItem(ctx context.Context, userID uint64, item domain.Item) error
	DeleteItem(ctx context.Context, userID uint64, sku domain.Sku) error
	DeleteItemsByUserID(ctx context.Context, userID uint64) error
	Checkout(ctx context.Context, userID uint64) (int64, error)
}

type validate interface {
	Struct(i any) error
}

type Server struct {
	cartService CartService
	validator   validate
}

func New(cartService CartService, validator validate) *Server {
	return &Server{
		cartService: cartService,
		validator:   validator,
	}
}

func (s *Server) InitRoutes() http.Handler {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("POST /user/{user_id}/cart/{sku_id}", s.AddItemHandler)
	http.HandleFunc("GET /user/{user_id}/cart", s.GetItemsByUserID)
	http.HandleFunc("DELETE /user/{user_id}/cart/{sku_id}", s.DeleteItemHandler)
	http.HandleFunc("DELETE /user/{user_id}/cart", s.DeleteCartByUserID)
	http.HandleFunc("POST /checkout/{user_id}", s.CheckoutHandler)

	h := middleware.NewLoggingMiddleware()(http.DefaultServeMux)
	h = middleware.HTTPMetrics(h)
	h = middleware.TracingMiddleware(h)

	return h
}
