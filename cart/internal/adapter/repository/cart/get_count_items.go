package cart

func (r *Repository) GetCountItems() uint32 {
	r.mx.RLock()
	defer r.mx.RUnlock()

	var total uint32
	for _, userCart := range r.cartByUserID {
		for _, item := range userCart {
			total += item.Count
		}
	}
	return total
}
