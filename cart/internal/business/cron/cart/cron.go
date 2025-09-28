package cart

type cartRepository interface {
	GetCountItems() uint32
}

type CronProcessor struct {
	cartRepository cartRepository
}

func New(cartRepository cartRepository) *CronProcessor {
	return &CronProcessor{
		cartRepository: cartRepository,
	}
}
