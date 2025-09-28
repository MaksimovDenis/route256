package cart

import "net/http"

type Item struct {
	Sku   uint64
	Count int
}

type RespWithData[T any] struct {
	HTTPResp *http.Response
	Data     T
}
