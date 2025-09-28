package product_service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"route256/cart/internal/domain"

	"github.com/opentracing/opentracing-go"
)

const (
	GetProductBySkuEndpoint = "http://%s/product/%d"
)

type Client struct {
	httpClient http.Client
	token      string
	address    string
}

type getProductResponseDTO struct {
	Name  string `json:"name"`
	Price uint32 `json:"price"`
	Sku   uint32 `json:"sku"`
}

func New(
	httpClient http.Client,
	token string,
	address string,
) *Client {
	return &Client{
		httpClient: httpClient,
		token:      token,
		address:    address,
	}
}

func (cl *Client) GetProductBySku(ctx context.Context, sku domain.Sku) (domain.Product, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productClient.GetProductBySku")
	defer span.Finish()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf(GetProductBySkuEndpoint, cl.address, sku),
		http.NoBody,
	)
	if err != nil {
		return domain.Product{}, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	req.Header.Add("X-API-KEY", cl.token)

	response, err := cl.httpClient.Do(req)
	if err != nil {
		return domain.Product{}, fmt.Errorf("httpClient.Do %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return domain.Product{}, domain.ErrProductNotFound
	}

	if response.StatusCode != http.StatusOK {
		return domain.Product{}, fmt.Errorf("productclient.GetProductBySku: %d", response.StatusCode)
	}

	var resp getProductResponseDTO
	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return domain.Product{}, fmt.Errorf("json.NewDecoder %w", err)
	}

	return domain.Product{
		Name:  resp.Name,
		Price: resp.Price,
		Sku:   domain.Sku(resp.Sku),
	}, nil
}
