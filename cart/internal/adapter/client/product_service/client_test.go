package product_service_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	productclient "route256/cart/internal/adapter/client/product_service"
	"route256/cart/internal/domain"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProductService_GetProductBySku(t *testing.T) {
	t.Parallel()

	handler := http.NewServeMux()
	handler.HandleFunc("/product/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-KEY") != "integration-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		skuStr := strings.TrimPrefix(r.URL.Path, "/product/")
		sku, err := strconv.Atoi(skuStr)
		require.NoError(t, err)

		if sku == 111 {
			product := domain.Product{
				Name:  "Integration Product",
				Price: 100,
				Sku:   111,
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(product)
			require.NoError(t, err)
			return
		}
		http.NotFound(w, r)
	})

	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	url := strings.TrimPrefix(ts.URL, "http://")
	service := productclient.New(
		*ts.Client(),
		"integration-token",
		url,
	)
	product, err := service.GetProductBySku(context.Background(), 111)
	require.NoError(t, err)

	fmt.Println(product)
}
