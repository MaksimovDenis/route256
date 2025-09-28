package lomsdata

import (
	model "route256/homework/e2e/clients/model/loms"
)

type OrderStatus string

var (
	TestOrder = &model.Order{
		Items: []model.Item{
			{
				Sku:   1076963,
				Count: 3,
			},
			{
				Sku:   135717466,
				Count: 2,
			},
		},
	}

	TestOrder2 = &model.Order{
		Items: []model.Item{
			{
				Sku:   135937324,
				Count: 1,
			},
		},
	}
)
