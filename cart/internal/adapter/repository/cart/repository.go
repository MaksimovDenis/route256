package cart

import (
	"route256/cart/internal/domain"
	"sync"
)

type cartByUserID = map[uint64]domain.ItemInfoByID

type Repository struct {
	cartByUserID cartByUserID
	mx           sync.RWMutex
}

func New(c int) *Repository {
	return &Repository{
		cartByUserID: make(cartByUserID, c),
	}
}
