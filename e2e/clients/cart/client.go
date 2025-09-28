package cart

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	model "route256/homework/e2e/clients/model/cart"
)

var (
	addItemEndpoint            = "%s/user/%d/cart/%d"
	getItemsByUserIDEndpoint   = "%s/user/%d/cart"
	deleteCartByUserIDEndpoint = "%s/user/%d/cart"
	deleteItemEndpoint         = "%s/user/%d/cart/%d"
	checkoutEndpoint           = "%s/checkout/%d"
)

type Client struct {
	baseURL string
	token   string
	client  *http.Client
}

func New(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		client:  http.DefaultClient,
	}
}

type addItemReq struct {
	Count int `json:"count"`
}

type Item struct {
	Sku   uint64 `json:"sku"`
	Name  string `json:"name"`
	Count int    `json:"count"`
	Price uint32 `json:"price"`
}

type CheckOutReponse struct {
	OrderID uint64 `json:"order_id"`
}

type GetCartResp struct {
	Items      []Item `json:"items"`
	TotalPrice uint32 `json:"total_price"`
}

func (c *Client) AddItem(ctx context.Context, userID uint64, item model.Item) (*http.Response, error) {
	url := fmt.Sprintf(addItemEndpoint, c.baseURL, userID, item.Sku)

	body, err := json.Marshal(addItemReq{Count: item.Count})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp, nil
}

func (c *Client) GetItemsByUserID(ctx context.Context, userID uint64) (model.RespWithData[GetCartResp], error) {
	url := fmt.Sprintf(getItemsByUserIDEndpoint, c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return model.RespWithData[GetCartResp]{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return model.RespWithData[GetCartResp]{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.RespWithData[GetCartResp]{
			HTTPResp: resp,
			Data:     GetCartResp{},
		}, nil
	}

	var items GetCartResp
	err = json.NewDecoder(resp.Body).Decode(&items)
	if err != nil {
		return model.RespWithData[GetCartResp]{}, err
	}

	return model.RespWithData[GetCartResp]{
		HTTPResp: resp,
		Data:     items,
	}, nil
}

func (c *Client) DeleteCartByUserID(ctx context.Context, userID uint64) (*http.Response, error) {
	url := fmt.Sprintf(deleteCartByUserIDEndpoint, c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp, nil
}

func (c *Client) DeleteItem(ctx context.Context, userID, sku uint64) (*http.Response, error) {
	url := fmt.Sprintf(deleteItemEndpoint, c.baseURL, userID, sku)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp, nil
}

func (c *Client) Checkout(ctx context.Context, userID uint64) (model.RespWithData[CheckOutReponse], error) {
	url := fmt.Sprintf(checkoutEndpoint, c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return model.RespWithData[CheckOutReponse]{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return model.RespWithData[CheckOutReponse]{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return model.RespWithData[CheckOutReponse]{
			HTTPResp: resp,
			Data:     CheckOutReponse{},
		}, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, body)
	}

	var orderID CheckOutReponse
	err = json.NewDecoder(resp.Body).Decode(&orderID)
	if err != nil {
		return model.RespWithData[CheckOutReponse]{}, err
	}

	return model.RespWithData[CheckOutReponse]{
		HTTPResp: resp,
		Data:     orderID,
	}, nil
}
