package domain

import "errors"

var (
	ErrProductNotFound     = errors.New("SKU должен существовать в сервисе")
	ErrIncorrectUserID     = errors.New("идентификатор пользователя должен быть натуральным числом (больше нуля)")
	ErrIncorrectSku        = errors.New("SKU должен быть натуральным числом (больше нуля)")
	ErrIncorrectCountValue = errors.New("количество должно быть натуральным числом (больше нуля)")
	ErrNotEnoughStocks     = errors.New("невозможно добавить товара по количеству больше, чем есть в стоках")
	ErrItemNotFound        = errors.New("товар в корзине не найден")

	ErrEmptyCart = errors.New("empty cart")
)
