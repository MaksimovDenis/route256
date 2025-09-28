package loms

import (
	"context"
	"fmt"
	"route256/cart/internal/domain"
	desc "route256/cart/internal/pb/loms/v1"
	"time"

	"github.com/opentracing/opentracing-go"
)

type Client struct {
	orderClient desc.OrdersClient
	stockClient desc.StocksClient
	timeout     time.Duration
}

func New(orderClient desc.OrdersClient, stockClient desc.StocksClient, timeout time.Duration) *Client {
	return &Client{
		orderClient: orderClient,
		stockClient: stockClient,
		timeout:     timeout,
	}
}

func (c *Client) OrderCreate(ctx context.Context, userID uint64, items []domain.CartItem) (int64, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "lomsClient.OrderCreate")
	defer span.Finish()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req := &desc.OrderCreateRequest{
		UserId: int64(userID), // #nosec G115
		Items:  itemsDomainToMap(items),
	}

	resp, err := c.orderClient.OrderCreate(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("orderClient.OrderCreate: %w", err)
	}

	return resp.OrderId, nil
}

func (c *Client) StocksInfo(ctx context.Context, sku uint64) (int64, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "lomsClient.StocksInfo")
	defer span.Finish()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req := &desc.StocksInfoRequest{
		Sku: int64(sku), // #nosec G115
	}

	resp, err := c.stockClient.StocksInfo(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("stockClient.StocksInfo: %w", err)
	}

	return int64(resp.Count), nil
}

func itemsDomainToMap(items []domain.CartItem) []*desc.Item {
	result := make([]*desc.Item, len(items))

	for idx, value := range items {
		result[idx] = &desc.Item{
			Sku:   int64(value.Item.Sku), // #nosec G115
			Count: value.Item.Count,
		}
	}

	return result
}
