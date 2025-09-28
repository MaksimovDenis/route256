package domain

type Sku uint64

type Product struct {
	Name  string
	Price uint32
	Sku   Sku
}
