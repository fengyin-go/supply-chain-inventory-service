package model

func CopyPurchaseItems(items []PurchaseItem) []PurchaseItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]PurchaseItem, len(items))
	copy(out, items)
	return out
}

func CopyPurchaseOrder(p *PurchaseOrder) *PurchaseOrder {
	if p == nil {
		return nil
	}
	cp := *p
	cp.Items = CopyPurchaseItems(p.Items)
	return &cp
}

func CopyInboundOrder(i *InboundOrder) *InboundOrder {
	if i == nil {
		return nil
	}
	cp := *i
	cp.Items = CopyPurchaseItems(i.Items)
	return &cp
}
