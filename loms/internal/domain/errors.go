package domain

import "errors"

var (
	ErrOrderNotFound           = errors.New("нет информации по данному заказу")
	ErrStockNotFound           = errors.New("нет информации о стоках по данному SKU")
	ErrNotEnoughStock          = errors.New("для всех товаров сток должен быть больше или равен запрашиваемому")
	ErrInvalidReserveOperation = errors.New("значение превышает количество остатков")
	ErrCancelOrder             = errors.New("невозможность отменить неудавшийся заказ, а также оплаченный")
	ErrPayStatusOrder          = errors.New("оплата заказа в невалидном статусе невозможна")
	ErrInternalServerError     = errors.New("проблемы из-за неисправностей в системе")
)
