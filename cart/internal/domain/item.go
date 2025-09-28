package domain

type ItemInfoByID map[Sku]Item

type Item struct {
	Sku   Sku
	Count uint32
}

type CartItem struct {
	Item
	Product
}
type Cart struct {
	Items      []CartItem
	TotalPrice uint32
}
