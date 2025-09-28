package cartdata

import model "route256/homework/e2e/clients/model/cart"

const (
	SkuBook = 1076963
	SkuChef = 1148162

	Token = "testToken"
)

var (
	ItemBook = model.Item{
		Sku:   SkuBook,
		Count: 2,
	}

	ItemChef = model.Item{
		Sku:   SkuChef,
		Count: 1,
	}
)
