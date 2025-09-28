package loms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	model "route256/homework/e2e/clients/model/loms"
	"strconv"
)

var (
	createOrderEndpoint  = "%s/order/create"
	getOrderInfoEndpoint = "%s/order/info?orderId=%s"
	payOrderEndpoint     = "%s/order/pay"
	cancelOrderEndpoint  = "%s/order/cancel"
	getStockInfo         = "%s/stock/info?sku=%d"
	getHealthCheck       = "%s/health"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client:  http.DefaultClient,
	}
}

type orderPayReq struct {
	OrderID string `json:"orderId"`
}

type orderCancelReq struct {
	OrderID string `json:"orderId"`
}

type createOrderResp struct {
	OrderID string `json:"orderID"`
}

type stockInfoResp struct {
	Count int64 `json:"count"`
}

type domainItem struct {
	Sku   string `json:"sku"`
	Count int64  `json:"count"`
}

type orderDomain struct {
	UserID string            `json:"userId"`
	Status model.OrderStatus `json:"status"`
	Items  []domainItem      `json:"items"`
}

type healthCheckResp struct {
	Message string `json:"message"`
}

func (cl *Client) CreateOrder(ctx context.Context, order model.Order) (model.RespWithData[createOrderResp], error) {
	url := fmt.Sprintf(createOrderEndpoint, cl.baseURL)

	domainItems := make([]domainItem, len(order.Items))
	for idx, value := range order.Items {
		domainItems[idx] = domainItem{
			Sku:   strconv.FormatInt(int64(value.Sku), 10),
			Count: value.Count,
		}
	}

	reqBody := orderDomain{
		UserID: strconv.FormatInt(int64(order.UserID), 10),
		Status: order.Status,
		Items:  domainItems,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return model.RespWithData[createOrderResp]{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return model.RespWithData[createOrderResp]{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := cl.client.Do(req)
	if err != nil {
		return model.RespWithData[createOrderResp]{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.RespWithData[createOrderResp]{
			HTTPResp: resp,
			Data:     createOrderResp{},
		}, nil
	}

	var orderID createOrderResp
	err = json.NewDecoder(resp.Body).Decode(&orderID)
	if err != nil {
		return model.RespWithData[createOrderResp]{}, err
	}

	return model.RespWithData[createOrderResp]{
		HTTPResp: resp,
		Data:     orderID,
	}, nil
}

func (cl *Client) OrderInfo(ctx context.Context, orderID string) (model.RespWithData[orderDomain], error) {
	url := fmt.Sprintf(getOrderInfoEndpoint, cl.baseURL, orderID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return model.RespWithData[orderDomain]{}, err
	}

	resp, err := cl.client.Do(req)
	if err != nil {
		return model.RespWithData[orderDomain]{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.RespWithData[orderDomain]{
			HTTPResp: resp,
			Data:     orderDomain{},
		}, nil
	}

	var result orderDomain
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return model.RespWithData[orderDomain]{}, err
	}

	return model.RespWithData[orderDomain]{
		HTTPResp: resp,
		Data:     result,
	}, nil
}

func (cl *Client) OrderPay(ctx context.Context, orderID string) (*http.Response, error) {
	url := fmt.Sprintf(payOrderEndpoint, cl.baseURL)

	body, err := json.Marshal(orderPayReq{OrderID: orderID})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := cl.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp, nil
}

func (cl *Client) OrderCancel(ctx context.Context, orderID string) (*http.Response, error) {
	url := fmt.Sprintf(cancelOrderEndpoint, cl.baseURL)

	body, err := json.Marshal(orderCancelReq{OrderID: orderID})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := cl.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp, nil
}

func (cl *Client) StockInfo(ctx context.Context, sku model.Sku) (model.RespWithData[stockInfoResp], error) {
	url := fmt.Sprintf(getStockInfo, cl.baseURL, sku)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return model.RespWithData[stockInfoResp]{}, err
	}

	resp, err := cl.client.Do(req)
	if err != nil {
		return model.RespWithData[stockInfoResp]{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.RespWithData[stockInfoResp]{
			HTTPResp: resp,
			Data:     stockInfoResp{},
		}, nil
	}

	var result stockInfoResp
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return model.RespWithData[stockInfoResp]{}, err
	}

	return model.RespWithData[stockInfoResp]{
		HTTPResp: resp,
		Data:     result,
	}, nil
}

func (cl *Client) HealthCheck(ctx context.Context) (model.RespWithData[healthCheckResp], error) {
	url := fmt.Sprintf(getHealthCheck, cl.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return model.RespWithData[healthCheckResp]{}, err
	}

	resp, err := cl.client.Do(req)
	if err != nil {
		return model.RespWithData[healthCheckResp]{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.RespWithData[healthCheckResp]{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result healthCheckResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return model.RespWithData[healthCheckResp]{}, err
	}

	return model.RespWithData[healthCheckResp]{
		HTTPResp: resp,
		Data:     result,
	}, nil
}
